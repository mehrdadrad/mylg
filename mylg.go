// myLG is command line looking glass that written with Go language
// it tries from its own icmp and external looking glasses tools
package main

import (
	"errors"
	"github.com/briandowns/spinner"
	"net"
	"regexp"
	"strings"
	"time"

	"github.com/mehrdadrad/mylg/cli"
	"github.com/mehrdadrad/mylg/icmp"
	"github.com/mehrdadrad/mylg/lg"
	"github.com/mehrdadrad/mylg/ns"
	"github.com/mehrdadrad/mylg/ripe"
)

// Provider is a interface to accessing lg
type Provider interface {
	Set(host, version string)
	GetDefaultNode() string
	GetNodes() []string
	ChangeNode(node string)
	Ping() (string, error)
}

var (
	// register looking glass hosts
	providers = map[string]Provider{"telia": new(lg.Telia), "level3": new(lg.Level3), "cogent": new(lg.Cogent)}
	pNames    = providerNames()
	nsr       *ns.Request
)

func providerNames() []string {
	pNames := []string{}
	for p := range providers {
		pNames = append(pNames, p)
	}
	return pNames
}

func validateProvider(p string) (string, error) {
	pNames := []string{}
	match, _ := regexp.MatchString("("+strings.Join(pNames, "|")+")", p)
	p = strings.ToLower(p)
	if match {
		return p, nil
	}
	return "", errors.New("provider not support")

}
func init() {
	// Initialize name server data
	nsr = ns.NewRequest()
	go nsr.Init()
}

func main() {
	var (
		err     error
		request string
		loop    = true
		cPName  = "local"
	)

	rep := make(chan string, 1)
	req := make(chan string, 1)
	nxt := make(chan struct{}, 1)

	c := cli.Init("local")
	go c.Run(req, nxt)

	r, _ := regexp.Compile(`(ping|trace|lg|ns|dig|asn|connect|node|local|mode|help|exit|quit)\s{0,1}(.*)`)
	s := spinner.New(spinner.CharSets[26], 220*time.Millisecond)

	for loop {
		select {
		case request, loop = <-req:
			if !loop {
				break
			}
			if len(request) < 1 {
				c.Next()
				continue
			}
			subReq := r.FindStringSubmatch(request)
			if len(subReq) == 0 {
				println("syntax error")
				c.Next()
				continue
			}
			prompt := c.GetPrompt()
			cmd := strings.TrimSpace(subReq[1])
			args := strings.TrimSpace(subReq[2])
			switch {
			case cmd == "ping" && cPName == "local":
				p := icmp.NewPing()
				ra, err := net.ResolveIPAddr("ip", args)
				if err != nil {
					println("cannot resolve", args, ": Unknown host")
					c.Next()
					continue
				}
				p.IP(ra.String())
				for n := 0; n < 4; n++ {
					p.Ping(rep)
					println(<-rep)
				}
				c.Next()
			case cmd == "ping":
				s.Prefix = "please wait "
				s.Start()
				providers[cPName].Set(args, "ipv4")
				m, _ := providers[cPName].Ping()
				s.Stop()
				println(m)
				c.Next()
			case cmd == "trace":
				trace := icmp.Trace{}
				trace.Run(args)
				c.Next()
			case cmd == "dig":
				nsr.Dig(args)
				c.Next()
			case cmd == "node":
				switch {
				case strings.HasPrefix(prompt, "lg"):
					if _, ok := providers[cPName]; ok {
						providers[cPName].ChangeNode(args)
						c.UpdatePromptN(args, 3)
					} else {
						println("the specified node doesn't support")
					}
				case strings.HasPrefix(prompt, "ns"):
					if !nsr.ChkNode(args) {
						println("error: argument is not valid")
						continue
					}
					c.UpdatePromptN(args, 3)
				}
				c.Next()
			case cmd == "local":
				cPName = "local"
				c.SetPrompt(cPName)
				c.Next()
			case cmd == "lg":
				c.SetPrompt("lg")
				c.UpdateCompleter("connect", pNames)
				c.Next()
			case cmd == "connect":
				switch {
				case strings.HasPrefix(prompt, "lg"):
					var pName string
					if pName, err = validateProvider(args); err != nil {
						println("provider not available")
						c.Next()
						continue
					}
					cPName = pName
					if _, ok := providers[cPName]; ok {
						c.UpdatePromptN(cPName+"/"+providers[cPName].GetDefaultNode(), 2)
						go func() {
							c.UpdateCompleter("node", providers[cPName].GetNodes())
						}()
					} else {
						println("it doesn't support")
					}
				case strings.HasPrefix(prompt, "ns"):
					if !nsr.ChkCountry(args) {
						println("error: argument is not valid")
						continue
					}
					c.SetPrompt("ns/" + args)
					nsr.SetNodeList(c)

				}
				c.Next()
			case cmd == "ns":
				nsr.SetCountryList(c)
				c.SetPrompt("ns")
				c.Next()
			case cmd == "asn":
				asn := ripe.ASN{Number: args}
				if asn.GetData() {
					asn.PrettyPrint()
				}
				c.Next()
			case cmd == "mode":
				if args == "vim" {
					c.SetVim()
				} else if args == "emacs" {
					c.SetEmacs()
				} else {
					println("the request mode doesn't support")
				}
				c.Next()
			case cmd == "help":
				c.Help()
				c.Next()
			case cmd == "exit", cmd == "quit":
				c.Close(nxt)
				close(req)
				// todo
			}
		}
	}
}
