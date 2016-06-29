package cli

import (
	"github.com/mehrdadrad/mylg/banner"
	"gopkg.in/readline.v1"
)

type Readline struct {
	instance *readline.Instance
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
