package main

import (
	"github.com/mehrdadrad/myping/cli"
	"github.com/mehrdadrad/myping/icmp"
	"github.com/mehrdadrad/myping/icmp/telia"
	"net"
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
		cProvider string = "local"
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
			switch {
			case subReq[1] == "ping" && cProvider == "local":
				p := icmp.NewPing()
				ra, err := net.ResolveIPAddr("ip", subReq[2])
				if err != nil {
					println("cannot resolve", subReq[2], ": Unknown host")
					nxt <- struct{}{}
					continue
				}
				p.IP(ra.String())
				for n := 0; n < 4; n++ {
					p.Ping(rep)
					println(<-rep)
				}
				nxt <- struct{}{}
			case subReq[1] == "ping" && cProvider == "telia":
				p := telia.Init(subReq[2], "ipv4", "Los Angeles")
				println(p.Ping())
				nxt <- struct{}{}
			case subReq[1] == "connect":
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
