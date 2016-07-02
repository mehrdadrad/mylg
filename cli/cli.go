package cli

import (
	"fmt"
	"github.com/mehrdadrad/mylg/banner"
	"gopkg.in/readline.v1"
	"strings"
)

type Readline struct {
	instance  *readline.Instance
	completer *readline.PrefixCompleter
	prompt    string
	next      chan struct{}
}

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

func (r *Readline) RemoveItemCompleter(pcItem string) {
	child := []readline.PrefixCompleterInterface{}
	for _, p := range r.completer.Children {
		if strings.TrimSpace(string(p.GetName())) != pcItem {
			child = append(child, p)
		}
	}

	r.completer.Children = child

}

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

func (r *Readline) SetPrompt(p string) {
	r.prompt = p
	r.instance.SetPrompt(p + "> ")
}

func (r *Readline) GetPrompt() string {
	return r.prompt
}

func (r *Readline) Refresh() {
	r.instance.Refresh()
}

func (r *Readline) SetVim() {
	if !r.instance.IsVimMode() {
		r.instance.SetVimMode(true)
		println("mode changed to vim")
	} else {
		println("mode already is vim")
	}
}

func (r *Readline) SetEmacs() {
	if r.instance.IsVimMode() {
		r.instance.SetVimMode(false)
		println("mode changed to emacs")
	} else {
		println("mode already is emacs")
	}
}

func (r *Readline) Next() {
	r.next <- struct{}{}
}

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

func (r *Readline) Close(next chan struct{}) {
	r.instance.Close()
}

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
