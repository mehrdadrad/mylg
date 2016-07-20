// Package cli provides all methods to control command line functions
package cli

import (
	"encoding/json"
	"fmt"
	"github.com/mehrdadrad/mylg/banner"
	"gopkg.in/readline.v1"
	"io/ioutil"
	"net/http"
	"regexp"
	"strconv"
	"strings"
)

// Readline structure
type Readline struct {
	instance  *readline.Instance
	completer *readline.PrefixCompleter
	prompt    string
	next      chan struct{}
}

// Init set readline imain items
func Init(prompt, version string) *Readline {
	var (
		r         Readline
		err       error
		completer = readline.NewPrefixCompleter(
			readline.PcItem("ping"),
			readline.PcItem("trace"),
			readline.PcItem("bgp"),
			readline.PcItem("connect"),
			readline.PcItem("node"),
			readline.PcItem("local"),
			readline.PcItem("lg"),
			readline.PcItem("ns"),
			readline.PcItem("dig"),
			readline.PcItem("whois"),
			readline.PcItem("scan"),
			readline.PcItem("peering"),
			readline.PcItem("help"),
			readline.PcItem("exit"),
		)
	)
	r.completer = completer
	r.instance, err = readline.NewEx(&readline.Config{
		Prompt:          prompt + "> ",
		HistoryFile:     "/tmp/myping",
		InterruptPrompt: "^C",
		EOFPrompt:       "exit",
		AutoComplete:    completer,
	})
	if err != nil {
		panic(err)
	}
	banner.Println(version) // print banner
	checkUpdate(version)    // check update version
	r.prompt = prompt       // init local prompt
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
	func() {
		for {
			line, err := r.instance.Readline()
			if err != nil { // io.EOF, readline.ErrInterrupt
				break
			}
			cmd <- line
			if _, ok := <-next; !ok {
				break
			}
		}
	}()
}

// Close the readline instance
func (r *Readline) Close(next chan struct{}) {
	r.instance.Close()
}

// Help print out the main help
func (r *Readline) Help() {
	fmt.Println(`Usage:
	The myLG tool developed to troubleshoot networking situations.
	The vi/emacs mode,almost all basic features is supported. press tab to see what options are available.

	connect <provider name>     connects to external looking glass, press tab to see the menu
	node <city/country name>    connects to specific node at current looking glass, press tab to see the available nodes
	local                       back to local
	lg                          change mode to extenal looking glass
	ns                          change mode to name server looking up
	ping                        ping ip address or domain name
	dig                         name server looking up
	whois                       resolve AS number/IP/CIDR to holder (provides by ripe ncc)
	scan                        scan tcp ports (you can provide range >scan host minport maxport)
	peering                     peering information (provides by peeringdb.com)
	`)
}

// checkUpdate checks if any new version is available
func checkUpdate(version string) {
	type mylg struct {
		Version string
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
		println(err.Error())
		return
	}
	if version != appCtl.Version {
		fmt.Printf("New version is available (v%s) mylg.io/download\n", appCtl.Version)
	}
}

//Flag parses the command arguments syntax:
// -flag=x
// -flag x
// help
func Flag(args string) (string, map[string]interface{}) {
	var (
		r   = make(map[string]interface{}, 10)
		err error
	)
	args = strings.TrimSpace(args)
	re := regexp.MustCompile(`(?i)-([a-z]+)={0,1}\s{0,1}([0-9|a-z|-]+)`)
	f := re.FindAllStringSubmatch(args, -1)
	for _, kv := range f {
		if len(kv) > 1 {
			r[kv[1]], err = strconv.Atoi(kv[2])
			if err != nil {
				r[kv[1]] = kv[2]
			}
			args = strings.Replace(args, kv[0], "", -1)
		}
	}
	if m, _ := regexp.MatchString(`(?i)help$`, args); m {
		r["help"] = true
	}
	args = strings.TrimSpace(args)
	return args, r
}

// SetFlag returns command option(s)
func SetFlag(flag map[string]interface{}, option string, v interface{}) interface{} {
	if sValue, ok := flag[option]; ok {
		switch v.(type) {
		case int:
			return sValue.(int)
		default:
			return sValue.(string)
		}
	} else {
		return v
	}
}
