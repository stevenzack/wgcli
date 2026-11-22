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
		return "\"" + s + "\""
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
