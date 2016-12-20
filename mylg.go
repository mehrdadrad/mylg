// myLG is an open source software utility which combines the functions
// of the different network probes in one network diagnostic tool.
package main

import (
	"errors"
	"fmt"
	"os"
	"os/exec"
	"regexp"
	"runtime"
	"strings"
	"sync"
	"time"

	"github.com/briandowns/spinner"

	"github.com/mehrdadrad/mylg/cli"
	"github.com/mehrdadrad/mylg/disc"
	"github.com/mehrdadrad/mylg/http/ping"
	"github.com/mehrdadrad/mylg/icmp"
	"github.com/mehrdadrad/mylg/lg"
	"github.com/mehrdadrad/mylg/nms"
	"github.com/mehrdadrad/mylg/ns"
	"github.com/mehrdadrad/mylg/packet"
	"github.com/mehrdadrad/mylg/peeringdb"
	"github.com/mehrdadrad/mylg/scan"
	"github.com/mehrdadrad/mylg/services/httpd"
	"github.com/mehrdadrad/mylg/speedtest"
	"github.com/mehrdadrad/mylg/whois"
)

const (
	version = "0.2.6"
)

// Provider represents looking glass
type Provider interface {
	Set(host, version string)
	GetDefaultNode() string
	GetNodes() []string
	ChangeNode(node string) bool
	Ping() (string, error)
	Trace() chan string
	BGP() chan string
}

var (
	pNames    = providerNames()
	req       = make(chan string, 1)
	nxt       = make(chan struct{}, 1)
	spin      = spinner.New(spinner.CharSets[26], 220*time.Millisecond)
	eArgs     = os.Args
	args      string
	prompt    string
	cPName    string
	noIf      bool = true
	cfg       cli.Config
	nmsClient nms.Client
	nsr       *ns.Request
	c         *cli.Readline

	// register looking glass hosts
	providers = map[string]Provider{
		"telia":  new(lg.Telia),  // telia
		"level3": new(lg.Level3), // level3
		"cogent": new(lg.Cogent), //cogent
		"ntt":    new(lg.NTT),    //cogent
	}

	// map cmd to function
	cmdFunc = map[string]func(){
		"web":       web,          // web dashboard
		"dump":      dump,         // dump traffic
		"disc":      discovery,    // network discovery
		"scan":      scanPorts,    // network scan
		"mode":      mode,         // editor mode
		"ping":      pingQuery,    // ping
		"trace":     trace,        // trace route
		"bgp":       BGP,          // BGP
		"whois":     whoisLookup,  // whois / dns lookup
		"peering":   peeringDB,    // peering DB
		"hping":     hping,        // hping
		"dig":       dig,          // dig
		"nms":       setNMS,       // network management system
		"node":      node,         // change node
		"connect":   connect,      // connect to a country or LG
		"local":     local,        // local
		"help":      help,         // help
		"exit":      cleanUp,      // clean up
		"quit":      cleanUp,      // clean up
		"show":      show,         // show config
		"set":       setConfig,    // set config
		"lg":        setLG,        // prepare looking glass
		"ns":        setNS,        // prepare name server
		"speedtest": speedTest,    // prepare name server
		"version":   printVersion, // prints version
	}
)

// init
func init() {
	// load configuration
	cfg = cli.LoadConfig()
	// initialize name server
	nsr = ns.NewRequest()
	go nsr.Init()
	// set current provider, prompt
	cPName = "local"
	prompt = "local"
	// with interface
	if len(eArgs) == 1 {
		// initialize cli
		c = cli.Init(version)
		go c.Run(req, nxt)
		// start web server
		go httpd.Run(cfg)
		// set interface enabled
		noIf = false
		// set local as default
		local()
	}

}

func main() {
	// command line w/o interface
	if noIf {
		cmd := eArgs[1]
		args = strings.Join(eArgs[2:], " ")
		if f, ok := cmdFunc[cmd]; ok {
			f()
		} else {
			println("Invalid command please try mylg help")
		}
		return
	}
	// command like w/ interface
LOOP:
	for {
		select {
		case request, ok := <-req:
			if !ok {
				break LOOP
			}
			if len(request) < 1 {
				c.Next()
				continue
			}
			subReq := cli.CMDRgx().FindStringSubmatch(request)
			if len(subReq) == 0 {
				println("syntax error")
				c.Next()
				continue
			}
			prompt = c.GetPrompt()
			args = strings.TrimSpace(subReq[2])
			cmd := strings.TrimSpace(subReq[1])
			if f, ok := cmdFunc[cmd]; ok {
				f()
			} else {
				println("Invalid command please try help")
			}
			c.Next()
		}
	}
}

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

// node handles node cmd
func node() {
	switch {
	case strings.HasPrefix(prompt, "lg"):
		if _, ok := providers[cPName]; ok {
			if providers[cPName].ChangeNode(args) {
				c.UpdatePromptN(args, 3)
				return
			}
		}
		println("the specified node doesn't support")
	case strings.HasPrefix(prompt, "ns"):
		if !nsr.ChkNode(args) {
			println("error: argument is not valid")
		} else {
			c.UpdatePromptN(args, 3)
		}
	default:
		if cPName == "local" {
			println("local doesn't support node")
		}

	}
}

// dig gets dig info
func dig() {
	if ok := nsr.SetOptions(args, prompt); ok {
		nsr.Dig()
	}
}

// web tries to open web interface at default web browser
func web() {
	var openCmd = "open"
	println("opening default web broswer ...")
	if runtime.GOOS != "darwin" {
		openCmd = "xdg-open"
	}
	cmd := exec.Command(openCmd, fmt.Sprintf("http://%s:%d", cfg.Web.Address, cfg.Web.Port))
	err := cmd.Start()
	if err != nil {
		println("error opening default browser")
	}

}

// dump provides decoding packets
func dump() {
	p, err := packet.NewPacket(args)
	if p == nil || err != nil {
		return
	}
	println(p.Banner())
	for l := range p.Open() {
		l.PrintPretty()
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

	case strings.HasPrefix(prompt, "nms"):
		nmsClient, err = nms.NewClient(args, cfg)
		if err != nil {
			println("error:", err.Error())
		} else if nmsClient.Host == "" {
			return
		} else {
			c.SetPrompt("nms/" + nmsClient.Host)
		}
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
		trace, err := icmp.NewTrace(args, cfg)
		if err != nil {
			println(err.Error())
		}
		if trace == nil {
			break
		}
		trace.Print()
	case strings.HasPrefix(prompt, "lg"):
		spin.Prefix = "please wait "
		spin.Start()
		providers[cPName].Set(args, "ipv4")
		for l := range providers[cPName].Trace() {
			if spin.Prefix != "" {
				spin.Stop()
				spin.Prefix = ""
				fmt.Printf("\n%s\n", l)
			} else {
				fmt.Println(l)
			}
		}
		spin.Stop()
	}
}

// hping tries to ping a web server by http
func hping() {
	// it should work at local mode
	if cPName != "local" {
		return
	}
	p, err := ping.NewPing(args, cfg)
	if err != nil {
		println(err.Error())
	} else {
		p.Run()
	}
}

// pingQuery runs ping command (local/LG)
func pingQuery() {
	if cPName == "local" {
		pingLocal()
	} else {
		pingLG()
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
	p, err := icmp.NewPing(args, cfg)
	if err != nil {
		println(err.Error())
	}
	if p == nil {
		return
	}
	if !p.IsCIDR() {
		resp := p.Run()
		p.PrintPretty(resp)
	} else {
		resp := p.MRun()
		p.CIDRHeader()
		for r := range resp {
			icmp.CIDRRespPrint(r)
		}
	}
}

func speedTest() {
	if err := speedtest.Run(); err != nil {
		println("\n", err.Error())
	}
}

// scanPorts tries to scan tcp/ip ports
func scanPorts() {
	scan, err := scan.NewScan(args, cfg)
	if err != nil {
		println(err.Error())
	} else {
		spin.Prefix = "please wait "
		spin.Start()
		scan.Run()
		spin.Stop()
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

// discovery handles disc command
func discovery() {
	var (
		wg sync.WaitGroup
		//ts = time.Now()
	)

	d := disc.New(args)
	// help requested
	if d == nil {
		return
	}

	spin.Prefix = "please wait "
	spin.Start()

	// load OUI async
	go func() {
		wg.Add(1)
		d.LoadOUI()
		wg.Done()
	}()

	d.PingLan()
	time.Sleep(5 * time.Second)

	if err := d.GetARPTable(); err != nil {
		println(err.Error())
		return
	}
	wg.Wait()
	spin.Stop()

	println("\nNetwork LAN Discovery")
	d.PrintPretty()
}

// nms
func setNMS() {
	items := []string{"interface", "config"}
	c.UpdateCompleter("show", items)
	c.SetPrompt("nms")
}

// setConfig
func setConfig() {
	if err := cli.SetConfig(args, &cfg); err != nil {
		println(err.Error())
	}
}

// show command
func show() {
	var err error

	re := regexp.MustCompile(`^([a-z]+)\s*(.*)$`)
	m := re.FindStringSubmatch(args)

	if len(m) < 3 {
		return
	}

	subItem := m[1]
	subArgs := m[2]

	switch subItem {
	case "config":
		cli.ShowConfig(&cfg)
	case "interface":
		if strings.HasPrefix(prompt, "nms") {
			err = nmsClient.ShowInterface(subArgs)
		} else {
			println("it's available under nms")
		}
		if err != nil {
			println(err.Error())
		}
	}
}

// setLG set lg prompt and completer
func setLG() {
	cPName = "telia"
	c.UpdateCompleter("connect", pNames)
	c.SetPrompt("lg/" + cPName + "/" + providers[cPName].GetDefaultNode())
	go func() {
		c.UpdateCompleter("node", providers[cPName].GetNodes())
	}()
}

// setNS set ns prompt and update completers
func setNS() {
	c.UpdateCompleter("connect", nsr.CountryList())
	c.UpdateCompleter("node", []string{})
	c.SetPrompt("ns")
}

// peeringDB gets peer info
func peeringDB() {
	peeringdb.Search(args)
}

// whoisLookup gets ANS/Prefix info
func whoisLookup() {
	whois.Lookup(args)
}

// local set prompts to local
func local() {
	nsr.Local()
	cPName = "local"
	c.UpdateCompleter("show", []string{"config"})
	c.SetPrompt(cPName)
}

// printVersion prints version and exits
func printVersion() {
	fmt.Printf("myLG v%s\n", version)
}

// cleanUp
func cleanUp() {
	c.Close(nxt)
	close(req)
}

// help
func help() {
	if noIf {
		// without command line
		h := `
              ***** TRY IT WITHOUT ANYTHING TO HAVE INTERFACE *****
        Usage:
              mylg [command] [args...]

              Available commands:

              ping                        ping ip address or domain name
              trace                       trace ip address or domain name (real-time w/ -r option)
              dig                         name server looking up
              whois                       resolve AS number/IP/CIDR to holder (provides by ripe ncc)
              hping                       Ping through HTTP/HTTPS w/ GET/HEAD methods
              scan                        scan tcp ports (you can provide range >scan host minport maxport)
              dump                        prints out a description of the contents of packets on a network interface
              disc                        discover all the devices on a LAN
              peering                     peering information (provides by peeringdb.com)
              version                     shows mylg version

        Example:
              mylg trace freebsd.org -r
              mylg whois 8.8.8.8
              mylg scan 127.0.0.1
              mylg dig google.com +trace
		`
		fmt.Println(h)
	} else {
		// with command line interface

		c.Help()
	}
}
