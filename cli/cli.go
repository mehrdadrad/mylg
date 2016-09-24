// Package cli provides all methods to control command line functions
package cli

import (
	"encoding/json"
	"fmt"
	"github.com/chzyer/readline"
	"github.com/mehrdadrad/mylg/banner"
	"io"
	"io/ioutil"
	"net/http"
	"regexp"
	"strconv"
	"strings"
)

const usage = `Usage:
	The myLG tool developed to troubleshoot networking situations.
	The vi/emacs mode,almost all basic features is supported. press tab to see what options are available.

	connect <provider name>     connects to external looking glass, press tab to see the menu
	node <city/country name>    connects to specific node at current looking glass, press tab to see the available nodes
	local                       back to local
	lg                          change mode to external looking glass
	ns                          change mode to name server looking up
	ping                        ping ip address or domain name
	trace                       trace ip address or domain name (real-time w/ -r option)
	dig                         name server looking up
	whois                       resolve AS number/IP/CIDR to holder (provides by ripe ncc)
	hping                       Ping through HTTP/HTTPS w/ GET/HEAD methods
	scan                        scan tcp ports (you can provide range >scan host minport maxport)
	dump                        prints out a description of the contents of packets on a network interface
	disc                        discover all the devices on a LAN                
	peering                     peering information (provides by peeringdb.com)
	web                         web dashboard - opens dashboard at your default browser
	`

// Readline structure
type Readline struct {
	instance  *readline.Instance
	completer *readline.PrefixCompleter
	prompt    string
	next      chan struct{}
}

var (
	cmds = []string{
		"ping",
		"trace",
		"bgp",
		"hping",
		"connect",
		"node",
		"local",
		"lg",
		"ns",
		"dig",
		"nms",
		"whois",
		"scan",
		"dump",
		"disc",
		"peering",
		"help",
		"web",
		"set",
		"exit",
		"show",
	}
)

// Init set readline imain items
func Init(version string) *Readline {
	var (
		r         Readline
		err       error
		completer = readline.NewPrefixCompleter(pcItems()...)
	)

	r.completer = completer
	r.instance, err = readline.NewEx(&readline.Config{
		Prompt:          "local> ",
		HistoryFile:     "/tmp/myping",
		InterruptPrompt: "^C",
		EOFPrompt:       "exit",
		AutoComplete:    completer,
	})
	if err != nil {
		panic(err)
	}

	banner.Println(version) // print banner
	go checkUpdate(version) // check update version

	return &r
}

// RemoveItemCompleter removes subitem(s) from a specific main item
func (r *Readline) RemoveItemCompleter(pcItem string) {
	child := []readline.PrefixCompleterInterface{}
	for _, p := range r.completer.Children {
		if strings.TrimSpace(string(p.GetName())) != pcItem {
			child = append(child, p)
		}
	}

	r.completer.Children = child

}

// AddCompleter updates subitem(s) from a specific main item
func (r *Readline) AddCompleter(pcItem string, pcSubItems []string) {
	var pc readline.PrefixCompleter
	c := []readline.PrefixCompleterInterface{}
	for _, item := range pcSubItems {
		c = append(c, readline.PcItem(item))
	}
	pc.Name = []rune(pcItem + " ")
	pc.Children = c
	r.completer.Children = append(r.completer.Children, &pc)
}

// UpdateCompleter updates subitem(s) from a specific main item
func (r *Readline) UpdateCompleter(pcItem string, pcSubItems []string) {
	child := []readline.PrefixCompleterInterface{}
	var pc readline.PrefixCompleter
	for _, p := range r.completer.Children {
		if strings.TrimSpace(string(p.GetName())) == pcItem {
			c := []readline.PrefixCompleterInterface{}
			for _, item := range pcSubItems {
				c = append(c, readline.PcItem(item))
			}
			pc.Name = []rune(pcItem + " ")
			pc.Children = c
			child = append(child, &pc)
		} else {
			child = append(child, p)
		}
	}
	if len(pc.Name) < 1 {
		// todo adding new
	}
	r.completer.Children = child
}

// SetPrompt set readline prompt and store it
func (r *Readline) SetPrompt(p string) {
	p = strings.ToLower(p)
	r.prompt = p
	r.instance.SetPrompt(p + "> ")
}

// UpdatePromptN appends readline prompt
func (r *Readline) UpdatePromptN(p string, n int) {
	var parts []string
	p = strings.ToLower(p)
	parts = strings.SplitAfterN(r.prompt, "/", n)
	if n <= len(parts) && n > -1 {
		parts[n-1] = p
		r.prompt = strings.Join(parts, "")
	} else {
		r.prompt += "/" + p
	}
	r.instance.SetPrompt(r.prompt + "> ")
}

// GetPrompt returns the current prompt string
func (r *Readline) GetPrompt() string {
	return r.prompt
}

// Refresh prompt
func (r *Readline) Refresh() {
	r.instance.Refresh()
}

// SetVim set mode to vim
func (r *Readline) SetVim() {
	if !r.instance.IsVimMode() {
		r.instance.SetVimMode(true)
		println("mode changed to vim")
	} else {
		println("mode already is vim")
	}
}

// SetEmacs set mode to emacs
func (r *Readline) SetEmacs() {
	if r.instance.IsVimMode() {
		r.instance.SetVimMode(false)
		println("mode changed to emacs")
	} else {
		println("mode already is emacs")
	}
}

// Next trigers to read next line
func (r *Readline) Next() {
	r.next <- struct{}{}
}

// Run the main loop
func (r *Readline) Run(cmd chan<- string, next chan struct{}) {
	r.next = next
	defer close(cmd)

LOOP:
	for {
		line, err := r.instance.Readline()
		if err != nil { // io.EOF, readline.ErrInterrupt
			switch err {
			case io.EOF:
				break LOOP
			case readline.ErrInterrupt:
			default:
				println(err.Error())
				break LOOP
			}
		}
		cmd <- line
		if _, ok := <-next; !ok {
			break
		}
	}
}

// Close the readline instance
func (r *Readline) Close(next chan struct{}) {
	r.instance.Close()
}

// Help print out the main help
func (r *Readline) Help() {
	fmt.Println(usage)
}

// CMDRex returns commands regex for validation
func CMDRgx() *regexp.Regexp {
	expr := fmt.Sprintf(`(%s)\s{0,1}(.*)`, strings.Join(cmds, "|"))
	re, _ := regexp.Compile(expr)
	return re
}

func pcItems() []readline.PrefixCompleterInterface {
	var (
		i        []readline.PrefixCompleterInterface
		subItems = map[string][]readline.PrefixCompleterInterface{
			"set": []readline.PrefixCompleterInterface{
				readline.PcItem("snmp",
					readline.PcItem("community"),
					readline.PcItem("version"),
					readline.PcItem("timeout"),
				),
				readline.PcItem("ping",
					readline.PcItem("interval"),
					readline.PcItem("count"),
					readline.PcItem("timeout"),
				),
				readline.PcItem("hping",
					readline.PcItem("method"),
					readline.PcItem("count"),
					readline.PcItem("timeout"),
					readline.PcItem("data"),
				),
				readline.PcItem("web",
					readline.PcItem("port"),
					readline.PcItem("address"),
				),
				readline.PcItem("scan",
					readline.PcItem("port"),
				),
				readline.PcItem("trace",
					readline.PcItem("wait"),
				),
			},
		}
	)

	for _, c := range cmds {
		if _, ok := subItems[c]; !ok {
			i = append(i, readline.PcItem(c))
		} else {
			i = append(i, readline.PcItem(c, subItems[c]...))
		}
	}
	return i
}

// checkUpdate checks if any new version is available
func checkUpdate(version string) {
	type mylg struct {
		Version string
		Update  struct {
			Enabled bool
		}
	}
	var appCtl mylg

	if version == "test" {
		return
	}

	resp, err := http.Get("http://mylg.io/appctl/mylg")
	if err != nil {
		println("error: check update has been failed ")
		return
	}
	defer resp.Body.Close()
	body, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		println("error: check update has been failed (2)" + err.Error())
		return
	}
	err = json.Unmarshal(body, &appCtl)
	if err != nil {
		return
	}
	if appCtl.Update.Enabled && version != appCtl.Version {
		fmt.Printf("New version is available (v%s) at http://mylg.io/download\n", appCtl.Version)
	}
}

//Flag parses the command arguments syntax:
// -flag=x
// -flag x
// help
func Flag(args string) (string, map[string]interface{}) {
	var (
		r      = make(map[string]interface{}, 10)
		err    error
		target string
	)

	// in case we have args without target
	args = " " + args

	// range
	re := regexp.MustCompile(`(?i)\s-([a-z|0-9]+)[=|\s]{0,1}(\d+\-\d+)`)
	f := re.FindAllStringSubmatch(args, -1)
	for _, kv := range f {
		r[kv[1]] = kv[2]
		args = strings.Replace(args, kv[0], "", 1)
	}
	// noon-boolean flags
	re = regexp.MustCompile(`(?i)\s{1}-([a-z|0-9]+)[=|\s]([0-9|a-z|'"{}:\/]+)`)
	f = re.FindAllStringSubmatch(args, -1)
	for _, kv := range f {
		if len(kv) > 1 {
			// trim extra characters (' and ") from value
			kv[2] = strings.Trim(kv[2], "'")
			kv[2] = strings.Trim(kv[2], `"`)
			r[kv[1]], err = strconv.Atoi(kv[2])
			if err != nil {
				r[kv[1]] = kv[2]
			}
			args = strings.Replace(args, kv[0], "", 1)
		}
	}
	// boolean flags
	re = regexp.MustCompile(`(?i)\s-([a-z|0-9]+)`)
	f = re.FindAllStringSubmatch(args, -1)
	for _, kv := range f {
		if len(kv) == 2 {
			r[kv[1]] = ""
			args = strings.Replace(args, kv[0], "", 1)
		}
	}
	// target
	re = regexp.MustCompile(`(?i)^[^-][\S|\w\s]*`)
	t := re.FindStringSubmatch(args)
	if len(t) > 0 {
		target = strings.TrimSpace(t[0])
	}
	// help
	if m, _ := regexp.MatchString(`(?i)help$`, args); m {
		r["help"] = true
	}

	return target, r
}

// SetFlag returns command option(s)
func SetFlag(flag map[string]interface{}, option string, v interface{}) interface{} {
	if sValue, ok := flag[option]; ok {
		switch v.(type) {
		case int:
			return sValue.(int)
		case string:
			switch sValue.(type) {
			case string:
				return sValue.(string)
			case int:
				str := strconv.FormatInt(int64(sValue.(int)), 10)
				return str
			case float64:
				str := strconv.FormatFloat(sValue.(float64), 'f', -1, 64)
				return str
			}
		case bool:
			return !v.(bool)
		default:
			return sValue.(string)
		}
	}
	return v
}
