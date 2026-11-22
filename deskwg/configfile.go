package main

import (
	"fmt"
	"log"

	"github.com/egoist/mygo"
	"github.com/stevenzack/wgcli/config"
)

func (a *app) importAccessKey(paths []string) (e error) {
	defer func() {
		if e != nil {
			mygo.Dialog.Error("导入失败", e.Error())
		}
	}()
	if len(paths) == 0 {
		return nil
	}
	p := paths[0]

	e = config.ImportAccessKeyFile(p)
	if e != nil {
		log.Println(e)
		return fmt.Errorf("导入AccessKey.csv文件失败:%w", e)
	}

	go a.initStatus()

	return nil
}
