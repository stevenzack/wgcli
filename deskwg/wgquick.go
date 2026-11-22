package main

import (
	"fmt"
	"log"
)

func wgQuickUp(conf string) error {
	output, e := RunAsAdmin("/opt/homebrew/bin/wg-quick", "up", conf)
	if e != nil {
		log.Println(e, output)
		return fmt.Errorf("%s: %w", output, e)
	}
	return nil
}

func wgQuickDown(conf string) error {
	output, e := RunAsAdmin("/opt/homebrew/bin/wg-quick", "down", conf)
	if e != nil {
		log.Println(e, output)
		return fmt.Errorf("%s: %w", output, e)
	}
	return nil
}
