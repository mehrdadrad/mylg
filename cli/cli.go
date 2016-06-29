package cli

import (
	"github.com/mehrdadrad/mylg/banner"
	"gopkg.in/readline.v1"
	"strings"
)

type Readline struct {
	instance  *readline.Instance
	completer *readline.PrefixCompleter
}

func Init(prompt string) *Readline {
	var (
		r         Readline
		err       error
		completer = readline.NewPrefixCompleter(
			readline.PcItem("ping"),
			readline.PcItem("connect",
				readline.PcItem("telia"),
				readline.PcItem("level3"),
			),
			readline.PcItem("node"),
			readline.PcItem("local"),
			readline.PcItem("help"),
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

func (r *Readline) UpdateCompleter(pcItem string, pcSubItems map[string]string) {
	child := []readline.PrefixCompleterInterface{}
	var pc readline.PrefixCompleter
	for _, p := range r.completer.Children {
		if strings.TrimSpace(string(p.GetName())) == pcItem {
			c := []readline.PrefixCompleterInterface{}
			for item, _ := range pcSubItems {
				c = append(c, readline.PcItem(item))
			}
			pc.Name = []rune("node ")
			pc.Children = c
			child = append(child, &pc)
		} else {
			child = append(child, p)
		}
	}

	r.completer.Children = child
}

func (r *Readline) SetPrompt(p string) {
	r.instance.SetPrompt(p + "> ")
}

func (r *Readline) Run(cmd chan<- string, next chan struct{}) {
	func() {
		for {
			line, err := r.instance.Readline()
			if err != nil { // io.EOF, readline.ErrInterrupt
				break
			}
			cmd <- line
			<-next
		}
	}()
}
