// Package cli provides all methods to control command line functions
package cli

import (
	"fmt"
	"github.com/mehrdadrad/mylg/banner"
	"gopkg.in/readline.v1"
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
func Init(prompt string) *Readline {
	var (
		r         Readline
		err       error
		completer = readline.NewPrefixCompleter(
			readline.PcItem("ping"),
			readline.PcItem("connect"),
			readline.PcItem("node"),
			readline.PcItem("local"),
			readline.PcItem("asn"),
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
	banner.Println()
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

	r.completer.Children = child
}

// SetPrompt set readline prompt and store it
func (r *Readline) SetPrompt(p string) {
	r.prompt = p
	r.instance.SetPrompt(p + "> ")
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
	ping                        ping ip address or domain name
	`)
}
