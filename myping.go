package main

import (
	"github.com/mehrdadrad/myping/cli"
	"github.com/mehrdadrad/myping/icmp"
	"regexp"
	"strings"
)

func validateProvider(p string) (string, error) {
	p = strings.ToLower(p)
	return p, nil
}

func main() {
	var (
		err       error
		cProvider string
	)

	rep := make(chan string, 1)
	cmd := make(chan string, 1)
	nxt := make(chan struct{}, 1)

	c := cli.Init("local")
	go c.Run(cmd, nxt)

	r, _ := regexp.Compile("(ping|connect) (.*)")

	for {
		select {
		case req := <-cmd:
			subReq := r.FindStringSubmatch(req)
			if len(subReq) == 0 {
				println("syntax error")
				nxt <- struct{}{}
				continue
			}
			switch subReq[1] {
			case "ping":
				p := icmp.NewPing()
				p.IP(subReq[2])
				p.Ping(rep)
				println(<-rep)
				nxt <- struct{}{}
			case "connect":
				var provider string
				if provider, err = validateProvider(subReq[2]); err != nil {
					println("provider not available")
					nxt <- struct{}{}
					continue
				}
				cProvider = provider
				c.SetPrompt(cProvider)
				nxt <- struct{}{}
			}
		}
	}
}
