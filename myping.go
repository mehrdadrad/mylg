package main

import (
	"github.com/mehrdadrad/myping/cli"
	"github.com/mehrdadrad/myping/icmp"
	"github.com/mehrdadrad/myping/icmp/telia"
	"net"
	"regexp"
	"strings"
)

type Provider interface {
	Init(host, version string)
	GetDefaultNode() string
	GetNodes() map[string]string
	Ping() (string, error)
}

var providers = map[string]Provider{"telia": new(telia.Provider)}

func validateProvider(p string) (string, error) {
	p = strings.ToLower(p)
	return p, nil
}

func main() {
	var (
		err    error
		cPName string = "local"
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
			case subReq[1] == "ping" && cPName == "local":
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
			case subReq[1] == "ping" && cPName == "telia":
				providers[cPName].Init(subReq[2], "ipv4")
				m, _ := providers[cPName].Ping()
				println(m)
				nxt <- struct{}{}
			case subReq[1] == "connect":
				var pName string
				if pName, err = validateProvider(subReq[2]); err != nil {
					println("provider not available")
					nxt <- struct{}{}
					continue
				}
				cPName = pName
				c.SetPrompt(cPName + "/" + providers[cPName].GetDefaultNode())
				nxt <- struct{}{}
			}
		}
	}
}
