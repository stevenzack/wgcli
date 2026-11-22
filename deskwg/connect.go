package main

import (
	"bytes"
	"fmt"
	"log"
	"strconv"

	"github.com/egoist/mygo"
	"github.com/stevenzack/wgcli/config"
	"github.com/stevenzack/wgcli/core"
)

func (a *app) doConnect() (e error) {
	defer func() {
		if e != nil {
			mygo.Dialog.Error("连接失败", e.Error())
		}
	}()
	//快速检查
	if a.region == "" {
		return fmt.Errorf("请选择翻墙目的地：比如新加坡")
	}
	region, ok := nameToRegion[a.region]
	if !ok {
		return fmt.Errorf("不支持的翻墙目的地：%s", a.region)
	}
	if a.hour == "" {
		return fmt.Errorf("必须选一个工作时长，比如3小时")
	}
	hour, e := strconv.Atoi(a.hour)
	if e != nil {
		log.Println(e)
		return e
	}
	if hour <= 0 {
		return fmt.Errorf("工作时长错误：%s", a.hour)
	}
	e = core.SetRegenName(region)
	if e != nil {
		log.Println(e)
		return e
	}

	a.status = Connecting

	a.log = new(bytes.Buffer)
	log.SetOutput(a)

	go func() {
		var status Status
		defer a.win.Update(func() {
			a.status = status
		})
		e := wgQuickDown(a.confPath)
		if e != nil {
			log.Println(e)
		}
		e = core.Deploy(hour, config.ConfigDir, func(path string) {
			e := wgQuickUp(path)
			if e != nil {
				log.Println(e)
				mygo.Dialog.Error("创建VPN连接失败", e.Error())
				status = Disconnected
				return
			}
			status = Connected
			a.confPath = path
		})
		if e != nil {
			log.Println(e)
			status = Disconnected
			mygo.Dialog.Error("部署失败：", e.Error())
			return
		}
	}()
	return nil
}

func (a *app) Write(bs []byte) (n int, e error) {
	if a.log != nil {
		n, e = a.log.Write(bs)
	} else {
		n = len(bs)
	}
	fmt.Print(string(bs))
	return n, e
}

func (a *app) doDisconnect() {
	go func() {
		e := wgQuickDown(a.confPath)
		if e != nil {
			log.Println(e)
			mygo.Dialog.Error("断开VPN连接失败", e.Error())
			a.win.Update(func() {
				a.status = Connected
			})
			return
		}
		a.win.Update(func() {
			a.status = Disconnected
		})
	}()
}
