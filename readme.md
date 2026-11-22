# README

[阿里云SDK文档](https://help.aliyun.com/zh/oss/developer-reference/manual-for-go-sdk-v2/)

# Program Execution Flow

- 检测AccessKey
- 导入access key
- 获取当前vpn状态
  - 开启VPN
    - 区域
    - duration
  - 打印日志
  - 保存.conf到appDir
  - wg-quick up /appDir/client.conf
  - 删除instance
- 定时检查连接状态 1/ms
  - 打印日志
  - 如果断联了，则暂时wg-quick down
    - 判断host是否到期by ICMP
    - 判断IP是否被封

# MyGo Demo

```go
ui.Text(c, "deskwg").FontSize(26).Bold()
ui.TextInput(c, &a.name).Placeholder("Your name").Label("Name").Width(240)
greeting := "Hello!"
if a.name != "" {
    greeting = fmt.Sprintf("Hello, %s!", a.name)
}
ui.Text(c, greeting).TextColor(t.TextMuted)
ui.Row(c).Gap(10).AlignItems(ui.Center).Children(func() {
    if ui.Button(c, "−").Width(44).Clicked() {
        a.count--
    }
    ui.Textf(c, "%d", a.count).FontSize(20).Width(48).TextAlign(ui.Center)
    if ui.PrimaryButton(c, "+").Width(44).Clicked() {
        a.count++
    }
})
```