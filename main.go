package main

import (
	"fmt"
	"log"

	_ "embed"

	"fyne.io/fyne/v2"
	"fyne.io/fyne/v2/app"
	"fyne.io/fyne/v2/container"
	"fyne.io/fyne/v2/data/binding"
	"fyne.io/fyne/v2/dialog"
	"fyne.io/fyne/v2/storage"
	"fyne.io/fyne/v2/theme"
	"fyne.io/fyne/v2/widget"
	"github.com/kirsle/configdir"
	"github.com/stevenzack/openurl"
	"github.com/stevenzack/wgcli/config"
	"github.com/stevenzack/wgcli/core"
)

var (
	a                            fyne.App
	mainWin                      fyne.Window
	logText                      string
	logTextBinding               = binding.BindString(&logText)
	deployLoading, deleteLoading bool
	deployLoadingBinding         = binding.BindBool(&deployLoading)
	deleteLoadingBinding         = binding.BindBool(&deleteLoading)
	accessKeyBinding             = binding.BindString(&config.AccessKeyID)
)

func init() {
	log.SetFlags(log.Lshortfile)
	// buf := new(bytes.Buffer)
	// log.SetOutput(buf)
	// go func() {
	// 	ticker := time.NewTicker(time.Second * 3)
	// 	for range ticker.C {
	// 		logText = buf.String()
	// 		loadingBinding.Reload()
	// 	}
	// }()
}
func main() {
	a = app.New()
	a.Settings().SetTheme(theme.DarkTheme())
	mainWin = a.NewWindow("WireGuard一键部署工具")
	mainWin.Resize(fyne.NewSize(600, 600))
	mainWin.SetContent(
		container.NewVBox(
			container.NewHBox(
				getDeployButton(),
				widget.NewButton("打开配置文件目录", func() {
					openurl.Open(configdir.LocalCache(config.AppName))
				}),
			),
			getDeleteButton(),
			widget.NewSeparator(),
			getAccessKeyComponent(),
			widget.NewButton("如何获取阿里云AccessKey?", func() {
				w := a.NewWindow("如何获取阿里云AccessKey")
				w.Resize(fyne.NewSize(400, 500))
				w.SetContent(widget.NewRichTextFromMarkdown(helpText))
				w.Show()
			}),
		),
	)
	mainWin.Show()

	a.Run()
}
func getDeleteButton() fyne.CanvasObject {
	bt := widget.NewButton("一键销毁WG服务器", func() {
		deleteLoading = true
		deleteLoadingBinding.Reload()
		go func() {
			e := core.Delete()
			if e != nil {
				log.Println(e)
				dialog.NewError(e, mainWin).Show()
			}
			deleteLoading = false
			deleteLoadingBinding.Reload()
		}()
	})
	act := widget.NewActivity()

	deleteLoadingBinding.AddListener(binding.NewDataListener(func() {
		b, e := deleteLoadingBinding.Get()
		if e == nil && b {
			bt.Disable()
			act.Show()
			act.Start()
		} else {
			bt.Enable()
			act.Stop()
			act.Hide()
		}
	}))
	return container.NewHBox(
		bt,
		act,
	)
}
func getDeployButton() fyne.CanvasObject {
	bt := widget.NewButton("一键部署WG服务器", func() {
		deployLoading = true
		deployLoadingBinding.Reload()
		go func() {
			e := core.Deploy()
			if e != nil {
				log.Println(e)
				dialog.NewError(e, mainWin).Show()
			}
			deployLoading = false
			deployLoadingBinding.Reload()
		}()
	})
	act := widget.NewActivity()

	deployLoadingBinding.AddListener(binding.NewDataListener(func() {
		b, e := deployLoadingBinding.Get()
		if e == nil && b {
			bt.Disable()
			act.Show()
			act.Start()
		} else {
			bt.Enable()
			act.Stop()
			act.Hide()
		}
	}))
	vbox := container.NewVBox(bt, act)
	return vbox
}

//go:embed helptext.md
var helpText string

func getAccessKeyComponent() fyne.CanvasObject {
	l := widget.NewLabel("")
	bt := widget.NewButton("导入AccessKey.csv", func() {
		dial := dialog.NewFileOpen(func(reader fyne.URIReadCloser, err error) {
			if err != nil {
				log.Println(err)
				return
			}
			reader.Close()
			fmt.Println("open file: ", reader.URI().String())
			e := config.ImportAccessKeyFile(reader.URI().Path())
			if e != nil {
				log.Println(e)
				return
			}
			accessKeyBinding.Reload()
		}, mainWin)
		dial.SetFilter(storage.NewExtensionFileFilter([]string{".csv"}))
		dial.SetTitleText("请选择从阿里云下载下来的AccessKey.csv文件")
		dial.Show()
	})
	accessKeyBinding.AddListener(binding.NewDataListener(func() {
		b, e := accessKeyBinding.Get()
		if e == nil && b != "" {
			l.SetText("AccessKey.csv已导入")
		} else {
			l.SetText("(尚未配置)")
		}
	}))
	return container.NewHBox(
		l,
		bt,
	)
}
