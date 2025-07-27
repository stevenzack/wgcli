package coreold

import (
	"fmt"
	"log"
	"os"
	"strings"
	"time"

	openapi "github.com/alibabacloud-go/darabonba-openapi/v2/client"
	ecs "github.com/alibabacloud-go/ecs-20140526/v7/client"
	"github.com/alibabacloud-go/tea/tea"
	credential "github.com/aliyun/credentials-go/credentials"
	"github.com/google/uuid"
	"github.com/pkg/sftp"
	"golang.org/x/crypto/ssh"
)

var (
	ID     = "xxx"
	Secret = "xxx"
)

const (
	defaultSecurityName = "sg-wgcli"
	defaultInstanceName = "ecs-wgcli"
	defaultRegionId     = "cn-hongkong"
	defaultPort         = "51823"
)

func FindAndCreateSecurityGroup() (sgId string, e error) {
	cli, e := CreateECSClient(ID, Secret)
	if e != nil {
		log.Println(e)
		return
	}
	res1, e := cli.DescribeSecurityGroups(&ecs.DescribeSecurityGroupsRequest{
		RegionId:          tea.String(defaultRegionId),
		SecurityGroupName: tea.String(defaultSecurityName),
	})
	if e != nil {
		log.Println(e)
		return
	}
	if len(res1.Body.SecurityGroups.SecurityGroup) > 0 {
		sgId = *res1.Body.SecurityGroups.SecurityGroup[0].SecurityGroupId
	} else {
		res, err := cli.CreateSecurityGroup(&ecs.CreateSecurityGroupRequest{
			SecurityGroupName: tea.String(defaultSecurityName),
			Description:       tea.String("Created by wgcli"),
			RegionId:          tea.String(defaultRegionId),
		})
		if err != nil {
			log.Println(err)
			e = err
			return
		}
		sgId = *res.Body.SecurityGroupId
	}
	fmt.Println("SG id=", sgId)

	_, e = cli.AuthorizeSecurityGroup(&ecs.AuthorizeSecurityGroupRequest{
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
		return
	}
	fmt.Println("OK")
	return
}
func DeleteInstance() {
	cli, e := CreateECSClient(ID, Secret)
	if e != nil {
		log.Println(e)
		return
	}
	instantId := "i-j6ca0jw9tsfek1fajwxh"
	res, e := cli.StopInstance(&ecs.StopInstanceRequest{
		InstanceId: tea.String(instantId),
	})
	if e != nil {
		log.Println(e)
		return
	}
	fmt.Println(res)
	time.Sleep(time.Second * 10)
	res2, e := cli.DeleteInstance(&ecs.DeleteInstanceRequest{
		InstanceId: tea.String(instantId),
	})
	if e != nil {
		log.Println(e)
		return
	}
	fmt.Println(res2)
}
func DownloadConf() {
	cli, e := ssh.Dial("tcp", "47.86.30.56:22", &ssh.ClientConfig{
		User: "root",
		Auth: []ssh.AuthMethod{
			ssh.Password(MD5(ID)[:10] + "A#"),
		},
		HostKeyCallback: ssh.InsecureIgnoreHostKey(),
	})
	if e != nil {
		log.Println(e)
		return
	}
	defer cli.Close()

	conn, e := sftp.NewClient(cli)
	if e != nil {
		log.Println(e)
		return
	}
	defer conn.Close()

}
func DialSSH() {
	cli, e := ssh.Dial("tcp", "47.86.30.56:22", &ssh.ClientConfig{
		User: "root",
		Auth: []ssh.AuthMethod{
			ssh.Password(MD5(ID)[:10] + "A#"),
		},
		HostKeyCallback: ssh.InsecureIgnoreHostKey(),
	})
	if e != nil {
		log.Println(e)
		return
	}
	defer cli.Close()
	fmt.Println("connected")
	e = runSession(cli, "apt update -y")
	if e != nil {
		log.Println(e)
		return
	}
	e = runSession(cli, "apt install wireguard -y")
	if e != nil {
		log.Println(e)
		return
	}
	e = runSession(cli, "mkdir -p /etc/wireguard")
	if e != nil {
		log.Println(e)
	}
	e = runSession(cli, "rm /etc/wireguard/*")
	if e != nil {
		log.Println(e)
	}

	e = runSession(cli, "wget -O /root/wg.sh https://get.vpnsetup.net/wg")
	if e != nil {
		log.Println(e)
		return
	}
	stdin := "\n51823\n\n\n\n"
	ses, e := cli.NewSession()
	if e != nil {
		log.Println(e)
		return
	}
	ses.RequestPty("bash", 600, 800, ssh.TerminalModes{})
	ses.Stdin = strings.NewReader(stdin)
	ses.Stdout = os.Stdout
	ses.Stderr = os.Stderr
	e = ses.Run("bash /root/wg.sh")
	if e != nil {
		log.Println(e)
		return
	}
	fmt.Println("OK")
}
func runSession(cli *ssh.Client, cmd string) error {
	ses, e := cli.NewSession()
	if e != nil {
		log.Println(e)
		return e
	}
	defer ses.Close()
	fmt.Println(cmd)
	e = ses.Run(cmd)
	if e != nil {
		log.Println(e)
		return e
	}
	return nil
}
func AllocateIP() {
	cli, e := CreateECSClient(ID, Secret)
	if e != nil {
		log.Println(e)
		return
	}
	instanceId := "i-j6ca0jw9tsfek1fajwxh"
	res, e := cli.AllocatePublicIpAddress(&ecs.AllocatePublicIpAddressRequest{
		InstanceId: &instanceId,
	})
	if e != nil {
		log.Println(e)
		return
	}
	fmt.Println(res)
}
func StartInstance() {
	cli, e := CreateECSClient(ID, Secret)
	if e != nil {
		log.Println(e)
		return
	}
	instanceId := "i-j6ca0jw9tsfek1fajwxh"
	req := &ecs.StartInstanceRequest{
		InstanceId: tea.String(instanceId),
	}
	res, e := cli.StartInstance(req)
	if e != nil {
		log.Println(e)
		return
	}
	fmt.Println(res)
}
func CreateInstance() {
	cli, e := CreateECSClient(ID, Secret)
	if e != nil {
		log.Println(e)
		return
	}
	res1, e := cli.DescribeInstances(&ecs.DescribeInstancesRequest{
		RegionId:     tea.String(defaultRegionId),
		InstanceName: tea.String(defaultInstanceName),
	})
	if e != nil {
		log.Println(e)
		return
	}

	var instanceId string
	if len(res1.Body.Instances.Instance) > 0 {
		instanceId = *res1.Body.Instances.Instance[0].InstanceId
	} else {
		ak, e := cli.GetAccessKeyId()
		if e != nil {
			log.Println(e)
			return
		}
		password := MD5(*ak)[:10] + "A#"
		fmt.Println(password)

		req := ecs.CreateInstanceRequest{
			RegionId:                tea.String("cn-hongkong"),
			AutoRenew:               tea.Bool(false),
			ClientToken:             tea.String(uuid.NewString()),
			InstanceType:            tea.String("ecs.e-c4m1.large"),
			InternetMaxBandwidthOut: tea.Int32(100),
			Password:                tea.String(password),
			SecurityGroupId:         tea.String("sg-j6c58apcnsg1qyght4vw"),
			SpotDuration:            tea.Int32(0),
			VSwitchId:               tea.String("vsw-j6cy87tejnarznm3u42ol"),
			ZoneId:                  tea.String("cn-hongkong-c"),
			InstanceName:            tea.String(defaultInstanceName),
			ImageId:                 tea.String("debian_12_8_x64_20G_alibase_20241216.vhd"),
			SystemDisk: &ecs.CreateInstanceRequestSystemDisk{
				Category: tea.String("cloud_essd_entry"),
				Size:     tea.Int32(20),
			},
			InstanceChargeType: tea.String("PostPaid"),
			SpotStrategy:       tea.String("SpotAsPriceGo"),
		}
		res, e := cli.CreateInstance(&req)
		if e != nil {
			log.Println(e)
			if strings.Contains(e.Error(), "InvalidAccountStatus.NotEnoughBalance") {

			}
			return
		}
		instanceId = *res.Body.InstanceId
	}

	fmt.Println("OK")
	fmt.Println(instanceId)
}

// Description:
//
// 使用凭据初始化账号Client
//
// @return Client
//
// @throws Exception
func CreateECSClient(ak, as string) (*ecs.Client, error) {
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
