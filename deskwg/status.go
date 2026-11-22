package main

import (
	"log"

	"github.com/stevenzack/wgcli/config"
)

func (a *app) initStatus() {
	//检测AccessKey状态
	e := config.LoadAccessKeyFile()
	if e != nil {
		log.Println(e)
		a.status = EmptyAccessKey
		return
	}
	a.status = Disconnected

	//检测连接状态
	if IsWireGuardConnected() {
		a.status = Connected
	}
}
