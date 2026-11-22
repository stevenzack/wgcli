package main

import (
	_ "embed"
	"flag"
	"fmt"
	"log"
	"os"
	"os/user"
	"path/filepath"
	"runtime"

	"github.com/stevenzack/openurl"
	"github.com/stevenzack/wgcli/config"
	"github.com/stevenzack/wgcli/core"
)

var (
	c     = flag.String("c", "", "import Aliyun AccessKey config file path (.csv)")
	del   = flag.Bool("d", false, "Delete existing instance")
	hour  = flag.Int("hour", 1, "Automatically delete it after X hours? (default 1 hour)")
	regen = flag.String("r", "hk", "regen, default is hk e.g. hk|sg|kr|jp")
	//go:embed helptext.md
	helpText string
)

func init() {
	log.SetFlags(log.Lshortfile)
}

func main() {
	flag.Parse()
	e := core.SetRegenName(*regen)
	if e != nil {
		log.Println(e)
		return
	}

	if *c != "" {
		log.Println("importing access key: ", *c)
		e := config.ImportAccessKeyFile(*c)
		if e != nil {
			log.Println(e)
			return
		}
	}

	log.Println("loading access key file")
	e = config.LoadAccessKeyFile()
	if e != nil {
		log.Println(e)

		if os.IsNotExist(e) {
			log.Println("ERROR: Aliyun access key is not configured")
			fmt.Println("如何获取阿里云AccessKey?")
			fmt.Println(helpText)
		}
		return
	}

	if *del {
		log.Println("deleting existing instance")
		e := core.Delete()
		if e != nil {
			log.Println(e)
			return
		}
		return
	}

	log.Println("deploying wireguard server...")
	dst, e := getDstDir()
	if e != nil {
		log.Println(e)
		return
	}
	e = core.Deploy(*hour, dst, func(path string) {
		openurl.Open(filepath.Dir(path))
	})
	if e != nil {
		log.Println(e)
		return
	}
	log.Println("OK")
}

func getDstDir() (string, error) {
	wd, e := user.Current()
	if e != nil {
		log.Println(e)
		return "", e
	}
	switch runtime.GOOS {
	// case "linux":
	// 	return "/etc/wireguard", nil
	default:
		return filepath.Join(wd.HomeDir, "Downloads"), nil
	}
}
