package cli

import (
	"gopkg.in/readline.v1"
)

type Readline struct {
	instance *readline.Instance
}

func Init(prompt string) *Readline {
	var (
		r   Readline
		err error
	)
	r.instance, err = readline.New(prompt + "> ")
	if err != nil {
		panic(err)
	}
	return &r
}

func (r *Readline) Run(out chan<- string) {
	func() {
		for {
			line, err := r.instance.Readline()
			if err != nil { // io.EOF, readline.ErrInterrupt
				break
			}
			out <- line
		}
	}()
}
