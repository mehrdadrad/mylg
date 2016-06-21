package ping

import (
	"golang.org/x/net/icmp"
)

type Ping struct {
	m icmp.Message
}

func (p *Ping) InitPing() {

}
