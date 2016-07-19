// Package ping tries to ping a HTTP server through different ways
// Connection, Session (Head), Get and Post
package ping

import (
	"fmt"
	"net"
	"net/url"
	"time"
)

// Ping represents HTTP ping request
type Ping struct {
	url     string
	timeout time.Duration
	rAddr   net.Addr
	conn    net.Conn
}

// NewPing validate and constructs request object
func NewPing(URL string, timeout time.Duration) *Ping {
	u, err := url.Parse(URL)
	if err != nil {

	}
	_, err = net.ResolveIPAddr("ip", u.Host)
	if err != nil {

	}
	return &Ping{
		url:     URL,
		timeout: timeout,
	}
}

// ping tries to create a TCP connection
func (p *Ping) ping() (float64, bool) {
	var (
		err   error
		delta float64
		sTime time.Time
	)

	sTime = time.Now()
	p.conn, err = net.DialTimeout("tcp", p.url, p.timeout*time.Second)
	delta = time.Since(sTime).Seconds()
	if err != nil {
		print(err.Error())
		return 0, false
	}

	p.rAddr = p.conn.RemoteAddr()
	p.conn.Close()
	return delta, true
}

// pingConn tries number of connection
func (p *Ping) pingConn() {
	for i := 0; i < 4; i++ {
		if t, ok := p.ping(); ok {
			fmt.Printf("HTTP Connection from %s seq=%d, time=%.3f ms\n", p.rAddr.String(), i, t*1000)
		} else {
			fmt.Println("HTTP Connection from %s seq=%d, timeout", p.rAddr.String(), i)
		}
	}
}

// pingSession tries to have multiple sessions
// per a connection
func (p *Ping) pingSession() {

}

// pingHead tries to execute head command
func (p *Ping) pingHead() {

}

func (p *Ping) Run() {
	p.pingConn()
}
