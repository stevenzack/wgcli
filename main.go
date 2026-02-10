package main

import (
	_ "embed"
	"flag"
	"fmt"
	"log"
	"os"

	"github.com/stevenzack/wgcli/config"
	"github.com/stevenzack/wgcli/core"
)

var (
	c     = flag.String("c", "", "import Aliyun AccessKey config file path (.csv)")
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
	e := core.SetRegen(*regen)
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

	log.Println("deploying wireguard server...")
	e = core.Deploy(*hour)
	if e != nil {
		log.Println(e)
		return
	}
	log.Println("OK")
}
