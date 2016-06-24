package main

import (
	"github.com/mehrdadrad/myping/cli"
	"github.com/mehrdadrad/myping/icmp"
	"regexp"
)

func main() {
	rep := make(chan string, 1)
	cmd := make(chan string, 1)
	nxt := make(chan struct{}, 1)

	c := cli.Init("local")
	go c.Run(cmd, nxt)

	r, _ := regexp.Compile("(ping|trace) (.*)")

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
				p.AddIP(subReq[2])
				p.Ping(rep)
				println(<-rep)
				nxt <- struct{}{}
			}
		}
	}
}
