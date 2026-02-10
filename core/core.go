package core

import (
	"errors"
	"io"
	"log"
	"os"
	"path/filepath"
	"strings"
	"time"

	openapi "github.com/alibabacloud-go/darabonba-openapi/v2/client"
	ecs "github.com/alibabacloud-go/ecs-20140526/v7/client"
	"github.com/alibabacloud-go/tea/tea"
	credential "github.com/aliyun/credentials-go/credentials"
	"github.com/google/uuid"
	"github.com/kirsle/configdir"
	"github.com/pkg/sftp"
	"github.com/stevenzack/openurl"
	"github.com/stevenzack/wgcli/config"
	"github.com/stevenzack/wgcli/utils"
	"golang.org/x/crypto/ssh"
)

const (
	defaultSecurityName = "sg-wgcli"
	defaultInstanceName = "ecs-wgcli"
	defaultRegionId     = "cn-hongkong"
	defaultZoneId       = "cn-hongkong-c"
	defaultPort         = "51823"
	defaultConf         = "client.conf"
)

func Delete() error {
	if config.AccessKeyID == "" || config.AccessKeySecret == "" {
		return errors.New("尚未导入阿里云AccessKey.csv文件，无法访问云服务")
	}
	// find instances
	cli, e := createEcsClient(config.AccessKeyID, config.AccessKeySecret)
	if e != nil {
		log.Println(e)
		return e
	}
	instanceId, ip, e := findInstance(cli)
	if e != nil {
		log.Println(e)
		return e
	}
	log.Println("stop instance: ", instanceId, ", ip: ", ip)
	_, e = cli.StopInstance(&ecs.StopInstanceRequest{
		InstanceId: &instanceId,
	})
	if e != nil {
		log.Println(e)
		return e
	}

	log.Println("delete instance")
	e = utils.Retry(3, time.Second*10, func() error {
		_, e := cli.DeleteInstance(&ecs.DeleteInstanceRequest{
			InstanceId: &instanceId,
		})
		if e != nil {
			log.Println(e)
			return e
		}
		return nil
	})
	if e != nil {
		log.Println(e)
		return e
	}

	log.Println("instance deleted")
	return nil
}
func Deploy(hour int) error {
	if config.AccessKeyID == "" || config.AccessKeySecret == "" {
		return errors.New("尚未导入阿里云AccessKey.csv文件，无法访问云服务")
	}
	// find instances
	cli, e := createEcsClient(config.AccessKeyID, config.AccessKeySecret)
	if e != nil {
		log.Println(e)
		return e
	}
	instanceId, ip, e := findInstance(cli)
	if e != nil {
		if e == os.ErrNotExist {
			//create instance
			instanceId, ip, e = createInstance(cli, hour)
			if e != nil {
				log.Println(e)
				return e
			}
		} else {
			log.Println(e)
			return e
		}
	}

	e = startInstance(cli, instanceId)
	if e != nil {
		log.Println(e)
		return e
	}

	e = utils.Retry(3, time.Second*15, func() error {
		e = dialSSH(ip)
		if e != nil {
			log.Println(e)
			return e
		}
		return nil
	})
	if e != nil {
		log.Println(e)
		return e
	}

	return nil
}

func dialSSH(ip string) error {
	log.Println("ssh connecting ", ip)
	cli, e := ssh.Dial("tcp", ip+":22", &ssh.ClientConfig{
		User: "root",
		Auth: []ssh.AuthMethod{
			ssh.Password(getPassword()),
		},
		HostKeyCallback: ssh.InsecureIgnoreHostKey(),
	})
	if e != nil {
		log.Println(e)
		return e
	}
	defer cli.Close()

	log.Println("ssh connected")
	e = utils.RunSession(cli, "apt update -y")
	if e != nil {
		log.Println(e)
		return e
	}
	e = utils.RunSession(cli, "apt install wireguard -y")
	if e != nil {
		log.Println(e)
		return e
	}

	e = utils.RunSession(cli, "mkdir -p /etc/wireguard")
	if e != nil {
		log.Println(e)
		return e
	}

	e = utils.RunSession(cli, "rm -rf /etc/wireguard/*")
	if e != nil {
		log.Println(e)
		return e
	}
	e = utils.RunSession(cli, "wget -O /root/wg.sh https://get.vpnsetup.net/wg")
	if e != nil {
		log.Println(e)
		return e
	}

	stdin := "\n51823\n\n\n\n"
	ses, e := cli.NewSession()
	if e != nil {
		log.Println(e)
		return e
	}
	ses.RequestPty("bash", 600, 800, ssh.TerminalModes{})
	ses.Stdin = strings.NewReader(stdin)
	ses.Stdout = os.Stdout
	ses.Stderr = os.Stderr
	log.Println("bash /root/wg.sh")
	e = ses.Run("bash /root/wg.sh")
	if e != nil {
		log.Println(e)
		return e
	}

	// download client.conf
	log.Println("now downloading client.conf")
	conn, e := sftp.NewClient(cli)
	if e != nil {
		log.Println(e)
		return e
	}
	if e != nil {
		log.Println(e)
		return e
	}
	defer conn.Close()

	dst := filepath.Join(configdir.LocalCache(config.AppName), defaultConf)
	e = os.MkdirAll(filepath.Dir(dst), 0755)
	if e != nil {
		log.Println(e)
		return e
	}
	fi, e := conn.Open("/root/client.conf")
	if e != nil {
		log.Println(e)
		return e
	}
	defer fi.Close()
	fo, e := os.OpenFile(dst, os.O_CREATE|os.O_TRUNC|os.O_WRONLY, 0644)
	if e != nil {
		log.Println(e)
		return e
	}
	defer fo.Close()
	_, e = io.Copy(fo, fi)
	if e != nil {
		log.Println(e)
		return e
	}

	log.Println("wireguard conf file saved at: ", dst)
	e = openurl.Open(filepath.Dir(dst))
	if e != nil {
		log.Println(e)
	}
	return nil
}
func startInstance(cli *ecs.Client, instId string) error {
	log.Println("start instance")
	_, e := cli.StartInstance(&ecs.StartInstanceRequest{
		InstanceId: tea.String(instId),
	})
	if e != nil {
		log.Println(e)
		return e
	}
	return nil
}

func createInstance(cli *ecs.Client, hour int) (id, ip string, err error) {
	log.Println("create instance")
	sgId, e := findAndCreateSecurityGroup(cli)
	if e != nil {
		log.Println(e)
		err = e
		return
	}
	vswId, e := findAndCreateVSwitch(cli)
	if e != nil {
		log.Println(e)
		err = e
		return
	}
	res, e := cli.CreateInstance(&ecs.CreateInstanceRequest{
		RegionId:                tea.String(defaultRegionId),
		AutoRenew:               tea.Bool(false),
		ClientToken:             tea.String(uuid.NewString()),
		InstanceType:            tea.String("ecs.e-c4m1.large"),
		InternetMaxBandwidthOut: tea.Int32(50),
		Password:                tea.String(getPassword()),
		SecurityGroupId:         &sgId,
		SpotDuration:            tea.Int32(0),
		VSwitchId:               tea.String(vswId),
		ZoneId:                  tea.String(defaultZoneId),
		InstanceName:            tea.String(defaultInstanceName),
		ImageId:                 tea.String("debian_12_8_x64_20G_alibase_20241216.vhd"),
		SystemDisk: &ecs.CreateInstanceRequestSystemDisk{
			Category: tea.String("cloud_essd_entry"),
			Size:     tea.Int32(20),
		},
		InstanceChargeType: tea.String("PostPaid"),
		SpotStrategy:       tea.String("SpotAsPriceGo"),
	})
	if e != nil {
		log.Println(e)
		err = e
		return
	}
	id = *res.Body.InstanceId

	// ip
	log.Println("allocate public ip address")
	res1, e := cli.AllocatePublicIpAddress(&ecs.AllocatePublicIpAddressRequest{
		InstanceId: tea.String(id),
	})
	if e != nil {
		log.Println(e)
		err = e
		return
	}
	ip = *res1.Body.IpAddress

	if hour > 0 {
		const layout = "2006-01-02T15:04:05Z"
		releaseTime := time.Now().Add(time.Hour * time.Duration(hour)).In(time.UTC).Format(layout)
		_, e := cli.ModifyInstanceAutoReleaseTime(&ecs.ModifyInstanceAutoReleaseTimeRequest{
			AutoReleaseTime: &releaseTime,
			InstanceId:      &id,
			RegionId:        cli.RegionId,
		})
		if e != nil {
			log.Println(e)
			err = e
			return
		}
		log.Println("set ecs instance auto release time to ", releaseTime," UTC")
	}
	return
}

func findAndCreateVSwitch(cli *ecs.Client) (swId string, err error) {
	log.Println("find and create vswitch")
	res, e := cli.DescribeVSwitches(&ecs.DescribeVSwitchesRequest{
		RegionId:  tea.String(defaultRegionId),
		ZoneId:    tea.String(defaultZoneId),
		IsDefault: tea.Bool(true),
	})
	if e != nil {
		log.Println(e)
		err = e
		return
	}
	l := res.Body.VSwitches.VSwitch
	if len(l) == 0 {
		err = errors.New("no vswitch found in zone " + defaultZoneId)
		return
	}
	swId = *l[0].VSwitchId
	return
}

func findAndCreateSecurityGroup(cli *ecs.Client) (sgId string, err error) {
	log.Println("find and create security group")
	l, e := cli.DescribeSecurityGroups(&ecs.DescribeSecurityGroupsRequest{
		RegionId:          tea.String(defaultRegionId),
		SecurityGroupName: tea.String(defaultSecurityName),
	})
	if e != nil {
		log.Println(e)
		err = e
		return
	}
	if len(l.Body.SecurityGroups.SecurityGroup) == 0 {
		res, e := cli.CreateSecurityGroup(&ecs.CreateSecurityGroupRequest{
			SecurityGroupName: tea.String(defaultSecurityName),
			RegionId:          tea.String(defaultRegionId),
			Description:       tea.String("Created by wgcli"),
		})
		if e != nil {
			log.Println(e)
			err = e
			return
		}
		sgId = *res.Body.SecurityGroupId
	} else {
		sg := l.Body.SecurityGroups.SecurityGroup[0]
		sgId = *sg.SecurityGroupId
	}

	e = authSecurityGroupPerm(cli, sgId)
	if e != nil {
		log.Println(e)
		err = e
		return
	}

	return
}
func authSecurityGroupPerm(cli *ecs.Client, sgId string) error {
	log.Println("auth security group permissions")
	_, e := cli.AuthorizeSecurityGroup(&ecs.AuthorizeSecurityGroupRequest{
		RegionId:        tea.String(defaultRegionId),
		SecurityGroupId: tea.String(sgId),
		Permissions: []*ecs.AuthorizeSecurityGroupRequestPermissions{
			{
				IpProtocol:   tea.String("UDP"),
				PortRange:    tea.String(defaultPort + "/" + defaultPort),
				Priority:     tea.String("1"),
				SourceCidrIp: tea.String("0.0.0.0/0"),
			},
			{
				IpProtocol:   tea.String("TCP"),
				PortRange:    tea.String("22/22"),
				Priority:     tea.String("1"),
				SourceCidrIp: tea.String("0.0.0.0/0"),
			},
		},
	})
	if e != nil {
		log.Println(e)
		return e
	}
	return nil
}
func getPassword() string {
	return utils.MD5(config.AccessKeyID)[:10] + "A#"
}
func findInstance(cli *ecs.Client) (id, ip string, err error) {
	log.Println("find instance")
	l, e := cli.DescribeInstances(&ecs.DescribeInstancesRequest{
		RegionId:     tea.String(defaultRegionId),
		InstanceName: tea.String(defaultInstanceName),
	})
	if e != nil {
		log.Println(e)
		err = e
		return
	}
	if len(l.Body.Instances.Instance) == 0 {
		err = os.ErrNotExist
		return
	}
	inst := l.Body.Instances.Instance[0]
	id = *inst.InstanceId

	if inst.PublicIpAddress == nil || len(inst.PublicIpAddress.IpAddress) == 0 {
		res, e := cli.AllocatePublicIpAddress(&ecs.AllocatePublicIpAddressRequest{
			InstanceId: tea.String(id),
		})
		if e != nil {
			log.Println(e)
			err = e
			return
		}
		ip = *res.Body.IpAddress
	} else {
		ip = *inst.PublicIpAddress.IpAddress[0]
	}

	return
}

// Description:
//
// 使用凭据初始化账号Client
//
// @return Client
//
// @throws Exception
func createEcsClient(ak, as string) (*ecs.Client, error) {
	log.Println("create ecs client")
	// 工程代码建议使用更安全的无AK方式，凭据配置方式请参见：https://help.aliyun.com/document_detail/378661.html。
	credential, e := credential.NewCredential(new(credential.Config).SetType("access_key").SetAccessKeyId(ak).SetAccessKeySecret(as))
	if e != nil {
		log.Println(e)
		return nil, e
	}

	config := &openapi.Config{
		Credential: credential,
	}
	// Endpoint 请参考 https://api.aliyun.com/product/Ecs
	config.Endpoint = tea.String("ecs.cn-hongkong.aliyuncs.com")
	cli, e := ecs.NewClient(config)
	if e != nil {
		log.Println(e)
		return nil, e
	}
	return cli, nil
}
