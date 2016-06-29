package main

import (
	"errors"
	"github.com/briandowns/spinner"
	"github.com/mehrdadrad/mylg/cli"
	"github.com/mehrdadrad/mylg/icmp"
	"github.com/mehrdadrad/mylg/lg"
	"net"
	"regexp"
	"strings"
	"time"
)

type Provider interface {
	Set(host, version string)
	GetDefaultNode() string
	GetNodes() map[string]string
	ChangeNode(node string)
	Ping() (string, error)
}

var providers = map[string]Provider{"telia": new(lg.Telia), "level3": new(lg.Level3)}

func validateProvider(p string) (string, error) {
	match, _ := regexp.MatchString("(telia|level3)", p)
	p = strings.ToLower(p)
	if match {
		return p, nil
	} else {
		return "", errors.New("provider not support")
	}
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

	r, _ := regexp.Compile(`(ping|connect|node|local) (.*)`)
	s := spinner.New(spinner.CharSets[26], 220*time.Millisecond)

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
			case subReq[1] == "ping":
				s.Prefix = "please wait "
				s.Start()
				providers[cPName].Set(subReq[2], "ipv4")
				m, _ := providers[cPName].Ping()
				s.Stop()
				println(m)
				nxt <- struct{}{}
			case subReq[1] == "node":
				providers[cPName].ChangeNode(subReq[2])
				c.SetPrompt(cPName + "/" + subReq[2])
				nxt <- struct{}{}
			case subReq[1] == "local":
				cPName = "local"
				c.SetPrompt(cPName)
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
