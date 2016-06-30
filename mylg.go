// myLG is command line looking glass that written with Go language
// it tries from its own icmp and external looking glasses tools
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
		err     error
		request string
		loop    bool   = true
		cPName  string = "local"
	)

	rep := make(chan string, 1)
	req := make(chan string, 1)
	nxt := make(chan struct{}, 1)

	c := cli.Init("local")
	go c.Run(req, nxt)

	r, _ := regexp.Compile(`(ping|connect|node|local|help|exit|quit)\s{0,1}(.*)`)
	s := spinner.New(spinner.CharSets[26], 220*time.Millisecond)

	for loop {
		select {
		case request, loop = <-req:
			if !loop {
				break
			}
			subReq := r.FindStringSubmatch(request)
			if len(subReq) == 0 {
				println("syntax error")
				nxt <- struct{}{}
				continue
			}
			cmd := strings.TrimSpace(subReq[1])
			args := strings.TrimSpace(subReq[2])
			switch {
			case cmd == "ping" && cPName == "local":
				p := icmp.NewPing()
				ra, err := net.ResolveIPAddr("ip", args)
				if err != nil {
					println("cannot resolve", args, ": Unknown host")
					nxt <- struct{}{}
					continue
				}
				p.IP(ra.String())
				for n := 0; n < 4; n++ {
					p.Ping(rep)
					println(<-rep)
				}
				nxt <- struct{}{}
			case cmd == "ping":
				s.Prefix = "please wait "
				s.Start()
				providers[cPName].Set(args, "ipv4")
				m, _ := providers[cPName].Ping()
				s.Stop()
				println(m)
				nxt <- struct{}{}
			case cmd == "node":
				if _, ok := providers[cPName]; ok {
					providers[cPName].ChangeNode(args)
					c.SetPrompt(cPName + "/" + args)
				} else {
					println("it doesn't support")
				}
				nxt <- struct{}{}
			case cmd == "local":
				cPName = "local"
				c.SetPrompt(cPName)
				nxt <- struct{}{}
			case cmd == "connect":
				var pName string
				if pName, err = validateProvider(args); err != nil {
					println("provider not available")
					nxt <- struct{}{}
					continue
				}
				cPName = pName
				if _, ok := providers[cPName]; ok {
					c.SetPrompt(cPName + "/" + providers[cPName].GetDefaultNode())
					go func() {
						c.UpdateCompleter("node", providers[cPName].GetNodes())
					}()
				} else {
					println("it doesn't support")
				}
				nxt <- struct{}{}
			case cmd == "mode":
				// todo
			case cmd == "help":
				c.Help()
				nxt <- struct{}{}
			case cmd == "exit", cmd == "quit":
				c.Close(nxt)
				close(req)
				// todo
			}
		}
	}
}
