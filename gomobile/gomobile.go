package gomobile

import (
	"io"
	"log"
	"os"

	"github.com/stevenzack/wgcli/config"
	"github.com/stevenzack/wgcli/core"
)

func init() {
	log.SetFlags(log.Lshortfile)
}

type CoreLogger io.Writer

func SetLogger(l CoreLogger) {
	log.SetOutput(l)
}

func SetRegionName(r string) {
	core.SetRegenName(r)
}

func SetConfigDir(dir string) {
	os.Setenv("XDG_CONFIG_HOME", dir)
	config.ConfigDir = dir
}
func SetCacheDir(dir string) {
	os.Setenv("XDG_CACHE_HOME", dir)
	config.CacheDir = dir
}

func GetRegionName() string {
	return core.GetRegionName()
}

func HasAccessKey() bool {
	return config.LoadAccessKeyFile() == nil
}

func ImportAccessKeyFile(file string) error {
	return config.ImportAccessKeyFile(file)
}

type FileHandler interface {
	OnConfFileSaved(path string)
}

func Deploy(hour int, confDstDir string, onOk FileHandler) error {
	log.Println("loading access key file")
	e := config.LoadAccessKeyFile()
	if e != nil {
		log.Println(e)
		return e
	}
	log.Println("deploying wireguard server...")
	e = core.Deploy(hour, confDstDir, func(path string) {
		onOk.OnConfFileSaved(path)
	})
	if e != nil {
		log.Println(e)
		return e
	}
	log.Println("OK")
	return nil
}
