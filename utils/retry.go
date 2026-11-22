package utils

import (
	"fmt"
	"log"
	"time"

	"golang.org/x/crypto/ssh"
)

func Retry(count int, interval time.Duration, action func() error) error {
	for i := 0; i < count; i++ {
		e := action()
		if e != nil {
			log.Println(e)
			log.Println("retry ", i+1, " times")
			time.Sleep(interval)
			continue
		}
		return nil
	}
	return fmt.Errorf("retried %d times and all failed", count)
}

func RunSession(cli *ssh.Client, cmd string) error {
	ses, e := cli.NewSession()
	if e != nil {
		log.Println(e)
		return e
	}
	defer ses.Close()
	log.Println(cmd)
	e = ses.Run(cmd)
	if e != nil {
		log.Println(e)
		return e
	}
	return nil
}
