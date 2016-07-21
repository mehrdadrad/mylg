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
	"github.com/mehrdadrad/mylg/http/ping"
	"github.com/mehrdadrad/mylg/icmp"
	"github.com/mehrdadrad/mylg/lg"
	"github.com/mehrdadrad/mylg/ns"
	"github.com/mehrdadrad/mylg/peeringdb"
	"github.com/mehrdadrad/mylg/ripe"
	"github.com/mehrdadrad/mylg/scan"
)

// Provider represents looking glass
type Provider interface {
	Set(host, version string)
	GetDefaultNode() string
	GetNodes() []string
	ChangeNode(node string)
	Ping() (string, error)
	Trace() chan string
	BGP() chan string
}

// Whois represents whois providers
type Whois interface {
	Set(r string)
	GetData() bool
	PrettyPrint()
}

const (
	version = "0.1.8"
)

var (
	// register looking glass hosts
	providers = map[string]Provider{"telia": new(lg.Telia), "level3": new(lg.Level3), "cogent": new(lg.Cogent)}
	whois     = map[string]Whois{"asn": new(ripe.ASN), "prefix": new(ripe.Prefix)}
	pNames    = providerNames()
	rep       = make(chan string, 1)
	req       = make(chan string, 1)
	nxt       = make(chan struct{}, 1)
	spin      = spinner.New(spinner.CharSets[26], 220*time.Millisecond)
	args      string
	prompt    string
	cPName    string
	nsr       *ns.Request
	c         *cli.Readline
)

// providerName
func providerNames() []string {
	pNames := []string{}
	for p := range providers {
		pNames = append(pNames, p)
	}
	return pNames
}

// validateProvider
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
	// initialize cli
	c = cli.Init("local", version)
	go c.Run(req, nxt)
	// initialize name server
	nsr = ns.NewRequest()
	go nsr.Init()
	// set default provider
	cPName = "local"
}

func main() {
	var (
		request string
		loop    = true
	)

	r, _ := regexp.Compile(`(ping|trace|bgp|lg|ns|dig|whois|peering|scan|hping|connect|node|local|mode|help|exit|quit)\s{0,1}(.*)`)

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
			prompt = c.GetPrompt()
			args = strings.TrimSpace(subReq[2])
			cmd := strings.TrimSpace(subReq[1])
			switch {
			case cmd == "hping" && cPName == "local":
				hping()
			case cmd == "ping" && cPName == "local":
				pingLocal()
			case cmd == "ping":
				pingLG()
			case cmd == "trace":
				trace()
			case cmd == "bgp":
				BGP()
			case cmd == "dig":
				nsr.Dig(args)
			case cmd == "node":
				node()
			case cmd == "local":
				nsr.Local()
				cPName = "local"
				c.SetPrompt(cPName)
			case cmd == "connect":
				connect()
			case cmd == "lg":
				c.SetPrompt("lg")
				c.UpdateCompleter("connect", pNames)
			case cmd == "ns":
				c.UpdateCompleter("connect", nsr.CountryList())
				c.UpdateCompleter("node", []string{})
				c.SetPrompt("ns")
			case cmd == "whois":
				whoIs()
			case cmd == "peering":
				peeringdb.Search(args)
			case cmd == "scan":
				scanPorts()
			case cmd == "mode":
				mode()
			case cmd == "help":
				c.Help()
			case cmd == "exit", cmd == "quit":
				c.Close(nxt)
				close(req)
			}
			// next line
			c.Next()
		}
	}
}

// node handles node cmd
func node() {
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
		} else {
			c.UpdatePromptN(args, 3)
		}
	}
}

// connect handles connect cmd
func connect() {
	var (
		pName string
		err   error
	)
	switch {
	case strings.HasPrefix(prompt, "lg"):
		if pName, err = validateProvider(args); err != nil {
			println("provider not available")
			c.Next()
			return
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
		} else {
			c.SetPrompt("ns/" + args)
			c.UpdateCompleter("node", nsr.NodeList())
		}
	}
}

// whoIs tries to get whois information
// ASN and prefix/ip information
func whoIs() {
	if ripe.IsASN(args) {
		whois["asn"].Set(args)
		whois["asn"].GetData()
		whois["asn"].PrettyPrint()
	} else {
		whois["prefix"].Set(args)
		whois["prefix"].GetData()
		whois["prefix"].PrettyPrint()
	}
}

// mode set editor mode
func mode() {
	if args == "vim" {
		c.SetVim()
	} else if args == "emacs" {
		c.SetEmacs()
	} else {
		println("the request mode doesn't support")
	}
}

// trace tries to trace from local and lg
func trace() {
	switch {
	case strings.HasPrefix(prompt, "local"):
		trace := icmp.Trace{}
		trace.Run(args)
	case strings.HasPrefix(prompt, "lg"):
		providers[cPName].Set(args, "ipv4")
		for l := range providers[cPName].Trace() {
			println(l)
		}
	}
}

// hping tries to ping a web server by http
func hping() {
	p, err := ping.NewPing(args, 5)
	if err != nil {
		println(err.Error())
	} else {
		p.Run()
	}
}

// pingLG tries to ping through a looking glass
func pingLG() {
	spin.Prefix = "please wait "
	spin.Start()
	providers[cPName].Set(args, "ipv4")
	m, err := providers[cPName].Ping()
	spin.Stop()
	if err != nil {
		println(err.Error())
	} else {
		println(m)
	}
}

// pingLocal tries to ping from local source ip
func pingLocal() {
	p := icmp.NewPing()
	ra, err := net.ResolveIPAddr("ip", args)
	if err != nil {
		println("cannot resolve", args, ": Unknown host")
		return
	}
	p.IP(ra.String())
	for n := 0; n < 4; n++ {
		p.Ping(rep)
		println(<-rep)
	}
}

// scanPorts tries to scan tcp/ip ports
func scanPorts() {
	scan, err := scan.NewScan(args)
	if err != nil {
		println(err.Error())
	} else {
		scan.Run()
	}
}

// BGP tries to get BGP lookup from a LG
func BGP() {
	if cPName == "local" {
		println("no provider selected")
		return
	}
	providers[cPName].Set(args, "ipv4")
	for l := range providers[cPName].BGP() {
		println(l)
	}
}
