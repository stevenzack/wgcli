package config

import (
	"bytes"
	"encoding/csv"
	"fmt"
	"log"
	"os"
	"path/filepath"

	"github.com/kirsle/configdir"
	"github.com/stevenzack/wgcli/utils"
)

const (
	AppName   = `wgcli`
	AppNameZh = "WireGuard一键部署工具"
)

var (
	AccessKeyID, AccessKeySecret string
	aesKey                       = []byte{163, 172, 28, 210, 169, 5, 152, 13, 83, 58, 114, 243, 56, 67, 168, 136, 132, 247, 151, 209, 9, 200, 17, 122, 196, 110, 167, 25, 151, 132, 245, 26}
)

func init() {
	e := LoadAccessKeyFile()
	if e != nil {
		log.Println(e)
	}
}

func getAccessKeyPath() (string, error) {
	dir := configdir.LocalConfig(AppName)
	e := configdir.MakePath(dir)
	if e != nil {
		log.Println(e)
		return "", e
	}
	const ak = `AccessKey.csv`
	dst := filepath.Join(dir, ak)
	return dst, nil
}

func ImportAccessKeyFile(csvPath string) error {
	b, e := os.ReadFile(csvPath)
	if e != nil {
		log.Println(e)
		return e
	}

	_, _, e = readCsvFile(b)
	if e != nil {
		log.Println(e)
		return e
	}

	b2, e := utils.EncryptAES(b, aesKey)
	if e != nil {
		log.Println(e)
		return e
	}

	dst, e := getAccessKeyPath()
	if e != nil {
		log.Println(e)
		return e
	}

	e = os.WriteFile(dst, b2, 0o600)
	if e != nil {
		log.Println(e)
		return e
	}

	e = LoadAccessKeyFile()
	if e != nil {
		log.Println(e)
		return e
	}

	return nil
}

func LoadAccessKeyFile() error {
	dst, e := getAccessKeyPath()
	if e != nil {
		log.Println(e)
		return e
	}

	b, e := os.ReadFile(dst)
	if e != nil {
		log.Println(e)
		return e
	}
	b2, e := utils.DecryptAES(b, aesKey)
	if e != nil {
		log.Println(e)
		return e
	}

	id, secret, e := readCsvFile(b2)
	if e != nil {
		log.Println(e)
		return e
	}
	AccessKeyID = id
	AccessKeySecret = secret
	return nil
}
func readCsvFile(b []byte) (id, secret string, err error) {
	r := csv.NewReader(bytes.NewReader(b))
	records, e := r.ReadAll()
	if e != nil {
		log.Println(e)
		err = e
		return
	}
	if len(records) != 2 || len(records[0]) != 2 || records[0][0] != "AccessKey ID" || records[0][1] != "AccessKey Secret" {
		err = fmt.Errorf("invalid record of AccessKey.csv: %v", records)
		return
	}
	id = records[1][0]
	secret = records[1][1]
	if id != "" && secret != "" {
		return
	}
	err = fmt.Errorf("invalid record of AccessKey.csv: %v", records)
	return
}
