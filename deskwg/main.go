package main

import (
	"bytes"
	"log"
	"os"
	"path/filepath"

	"github.com/egoist/mygo"
	"github.com/egoist/mygo/ui"
	"github.com/stevenzack/wgcli/config"
	"github.com/stevenzack/wgcli/core"
)

// app is the state the window shows. Its view builds the interface from
// it, on the main thread, whenever the window needs a frame: after input,
// after Window.Update, and while something animates.
type app struct {
	win            *mygo.Window
	status         Status
	region         string
	hour           string
	log            *bytes.Buffer
	confPath       string
	toDeleteRegion string
}

func (a *app) view(c *ui.Context) {
	t := c.Theme()
	ui.Column(c).Fill().Center().Gap(16).Padding(24).Children(func() {
		switch a.status {
		case EmptyAccessKey:
			ui.Text(c, "请导入阿里云的AccessKey.csv配置文件").TextColor(t.TextMuted)
			if ui.Button(c, "选择.csv文件").Clicked() {
				paths, e := mygo.Dialog.Open(mygo.OpenDialogOptions{
					Title:    "请选择阿里云AccessKey.csv文件",
					Multiple: false,
					Filters: []mygo.FileFilter{
						{
							Name:       "csv文件",
							Extensions: []string{"csv"},
						},
					},
				})
				if e != nil {
					log.Println(e)
				} else {
					a.importAccessKey(paths)
				}
			}
		case Disconnected:
			ui.Text(c, "选择翻墙目的地:").TextColor(t.TextMuted)
			ui.Select(c, &a.region, []string{
				Singapore,
				HongKong,
				Korea,
			})
			ui.Text(c, "你准备工作多少个小时？").TextColor(t.TextMuted)
			ui.Select(c, &a.hour, []string{
				"1",
				"3",
				"8",
			})
			if ui.PrimaryButton(c, "开始翻墙").Clicked() {
				a.doConnect()
			}
		case Connecting:
			ui.Scroll(c).Center().Fill().Children(func() {
				ui.Text(c, a.log.String()).Fill().TextColor(t.TextMuted)
			})
			ui.Spinner(c)
		case Connected:
			ui.Text(c, "已连接").TextColor(t.TextMuted)
			if ui.PrimaryButton(c, "断开连接").Clicked() {
				a.doDisconnect()
			}
			ui.Text(c, "选择要删除的目的地:").TextColor(t.TextMuted)
			ui.Select(c, &a.toDeleteRegion, []string{
				Singapore,
				HongKong,
				Korea,
			})
			if ui.Button(c, "删除节点").Clicked() {
				if a.toDeleteRegion == "" {
					mygo.Dialog.Error("删除节点失败", "必须先选择要删除的目的地")
					return
				}
				a.status = Deleting
				go func() {
					log.SetOutput(a)
					e := wgQuickDown(a.confPath)
					if e != nil {
						log.Println(e)
						mygo.Dialog.Error("删除失败：", e.Error())
					} else {
						core.SetRegenName(nameToRegion[a.toDeleteRegion])
						e := core.Delete()
						if e != nil {
							log.Println(e)
							mygo.Dialog.Error("删除失败：", e.Error())
						} else {
							os.Remove(a.confPath)
						}
					}
					a.initStatus()
					a.win.Update(func() {})
				}()
			}
		case Deleting:
			ui.Scroll(c).Center().Fill().Children(func() {
				ui.Text(c, a.log.String()).Fill().TextColor(t.TextMuted)
			})
			ui.Spinner(c)

		}
	})
}
func init() {
	log.SetFlags(log.Lshortfile)
}

func main() {
	a := &app{
		confPath: filepath.Join(config.ConfigDir, "client.conf"),
	}
	a.initStatus()
	mygo.App.WhenReady(func() {
		a.win = mygo.NewWindow(mygo.WindowOptions{
			Title:     "deskwg",
			Width:     480,
			Height:    360,
			MinWidth:  320,
			MinHeight: 280,
			// Opens where the user left it last time.
			StateKey: "main",
			// The window shows the interface MyGo draws, not a web page.
			Content: ui.View(a.view),
		})
	})
	if err := mygo.App.Run(); err != nil {
		log.Fatal(err)
	}
}
