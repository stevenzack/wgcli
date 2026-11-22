package main

import (
	"fmt"
	"log"
	"os/exec"
	"strings"
)

// /opt/homebrew/bin/wg-quick up /etc/wireguard/wg0.conf
func RunAsAdmin(name string, args ...string) (string, error) {
	// 正确处理 shell 参数中的引号和特殊字符
	quote := func(s string) string {
		return "'" + strings.ReplaceAll(s, "'", "'\\''") + "'"
	}

	parts := []string{quote(name)}
	for _, arg := range args {
		parts = append(parts, quote(arg))
	}

	shellCmd := strings.Join(parts, " ")

	// AppleScript 字符串转义
	appleQuote := func(s string) string {
		s = strings.ReplaceAll(s, "\\", "\\\\")
		s = strings.ReplaceAll(s, "\"", "\\\"")
		return "\"export PATH=/opt/homebrew/bin:$PATH; " + s + "\""
	}

	script := "do shell script " +
		appleQuote(shellCmd) +
		" with administrator privileges"

	cmd := exec.Command(
		"/usr/bin/osascript",
		"-e",
		script,
	)
	log.Println(cmd.String())
	output, err := cmd.CombinedOutput()
	if err != nil {
		return "", fmt.Errorf(
			"admin command failed: %w: %s",
			err, output,
		)
	}
	log.Println("output: ", string(output))
	return string(output), nil
}

/*
utun0: flags=8051<UP,POINTOPOINT,RUNNING,MULTICAST> mtu 1500

	inet6 fe80::81b1:e135:9e82:affc%utun0 prefixlen 64 scopeid 0xf
	nd6 options=201<PERFORMNUD,DAD>

utun1: flags=8051<UP,POINTOPOINT,RUNNING,MULTICAST> mtu 1380

	inet6 fe80::4cfc:b6f6:5cf9:ebed%utun1 prefixlen 64 scopeid 0x10
	nd6 options=201<PERFORMNUD,DAD>

utun2: flags=8051<UP,POINTOPOINT,RUNNING,MULTICAST> mtu 2000

	inet6 fe80::2c90:fbdd:d408:396e%utun2 prefixlen 64 scopeid 0x11
	nd6 options=201<PERFORMNUD,DAD>

utun3: flags=8051<UP,POINTOPOINT,RUNNING,MULTICAST> mtu 1000

	inet6 fe80::ce81:b1c:bd2c:69e%utun3 prefixlen 64 scopeid 0x12
	nd6 options=201<PERFORMNUD,DAD>

nan0: flags=8822<BROADCAST,SMART,SIMPLEX,MULTICAST> mtu 1500

	options=400<CHANNEL_IO>
	ether c2:eb:74:20:fb:83
	media: autoselect (none)

utun4: flags=8051<UP,POINTOPOINT,RUNNING,MULTICAST> mtu 1420

	inet 10.7.0.2 --> 10.7.0.2 netmask 0xffffff00
*/
func IsWireGuardConnected() bool {
	cmd := exec.Command("/sbin/ifconfig")
	log.Println(cmd.String())
	output, e := cmd.CombinedOutput()
	if e != nil {
		log.Println(e)
		return false
	}
	for _, line := range strings.Split(string(output), "\n") {
		if strings.Contains(line, "-->") {
			return true
		}
	}
	return false
}
