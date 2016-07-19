// Package ping tries to ping a HTTP server through different ways
// Connection, Session (Head), Get and Post
package ping

import (
	"bufio"
	"fmt"
	"net"
	"net/url"
	"reflect"
	"regexp"
	"time"
)

// Ping represents HTTP ping request
type Ping struct {
	url     string
	host    string
	timeout time.Duration
	rAddr   net.Addr
	conn    net.Conn
}

type Result struct {
	StatusCode int
	ConnTime   float64
	TotalTime  float64
	Proto      string
	Server     string
	Status     string
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
		host:    u.Host,
		timeout: timeout,
	}
}

// ping tries to create a TCP connection
func (p *Ping) pingConn() (Result, bool) {
	var (
		r     Result
		err   error
		sTime time.Time
	)

	sTime = time.Now()
	p.conn, err = net.DialTimeout("tcp", p.url, p.timeout*time.Second)
	r.ConnTime = time.Since(sTime).Seconds()
	if err != nil {
		print(err.Error())
		return r, false
	}

	p.rAddr = p.conn.RemoteAddr()
	p.conn.Close()
	return r, true
}

// pingConnLoop tries number of connection
func (p *Ping) pingConnLoop() {
	for i := 0; i < 4; i++ {
		if r, ok := p.pingConn(); ok {
			fmt.Printf("HTTP Connection from %s seq=%d, time=%.3f ms\n", p.rAddr.String(), i, r.ConnTime*1000)
		} else {
			fmt.Println("HTTP Connection from %s seq=%d, timeout", p.rAddr.String(), i)
		}
	}
}

// pingHeadLoop tries number of connection
// with header information
func (p *Ping) pingHeadLoop() {
	pStrPrefix := "HTTP connection to %s seq=%d, "
	pStrPostfix := "proto=%s, status=%s, time=%.3f ms\n"
	for i := 0; i < 4; i++ {
		if r, ok := p.pingHead(); ok {
			fmt.Printf(pStrPrefix+pStrPostfix, p.rAddr.String(), i, r.Proto, r.Status, r.TotalTime*1000)
		} else {
			fmt.Println(pStrPrefix+"timeout", p.rAddr.String(), i)
		}
	}
}

// pingSession tries to have multiple sessions
// per a connection
func (p *Ping) pingSession() {

}

// pingHead tries to execute head command
func (p *Ping) pingHead() (Result, bool) {
	var (
		r     Result
		b     = make([]byte, 512)
		err   error
		sTime time.Time
	)

	sTime = time.Now()
	p.conn, err = net.DialTimeout("tcp", p.host+":80", p.timeout*time.Second)

	if err != nil {
		print(err.Error())
		return r, false
	}

	fmt.Fprintf(p.conn, "HEAD / HTTP/1.1\r\n\r\n")
	reader := bufio.NewReader(p.conn)
	n, _ := reader.Read(b)
	for key, regex := range map[string]string{"Proto": `(HTTP/\d\.\d)`, "Status": `HTTP/\d\.\d\s+(\d+)`, "Server": `server:\s+(.*)\n`} {
		re := regexp.MustCompile(regex)
		a := re.FindSubmatch(b[:n])
		if len(a) == 2 {
			f := reflect.ValueOf(&r).Elem().FieldByName(key)
			f.SetString(string(a[1]))
		}
	}
	r.TotalTime = time.Since(sTime).Seconds()

	p.rAddr = p.conn.RemoteAddr()
	p.conn.Close()
	return r, true

}

func (p *Ping) Run() {
	p.pingHeadLoop()
}
