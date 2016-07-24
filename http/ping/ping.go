// Package ping tries to ping a HTTP server through different ways
// Connection, Session (Head), Get and Post
package ping

import (
	"fmt"
	"io/ioutil"
	"net"
	"net/http"
	"net/url"
	"regexp"
	"strings"
	"time"

	"github.com/mehrdadrad/mylg/cli"
)

// Ping represents HTTP ping request
type Ping struct {
	url     string
	host    string
	timeout time.Duration
	count   int
	method  string
	buf     string
	rAddr   net.Addr
	nsTime  time.Duration
	conn    net.Conn
}

// Result holds Ping result
type Result struct {
	StatusCode int
	ConnTime   float64
	TotalTime  float64
	Size       int
	Proto      string
	Server     string
	Status     string
}

// NewPing validate and constructs request object
func NewPing(args string) (*Ping, error) {
	URL, flag := cli.Flag(args)
	// help
	if _, ok := flag["help"]; ok || URL == "" {
		help()
		return nil, fmt.Errorf("")
	}
	URL = Normalize(URL)
	u, err := url.Parse(URL)
	if err != nil {
		return &Ping{}, fmt.Errorf("cannot parse url")
	}
	sTime := time.Now()
	ipAddr, err := net.ResolveIPAddr("ip", u.Host)
	if err != nil {
		return &Ping{}, fmt.Errorf("cannot resolve %s: Unknown host", u.Host)
	}

	p := &Ping{
		url:    URL,
		host:   u.Host,
		rAddr:  ipAddr,
		nsTime: time.Since(sTime),
	}

	// set count
	p.count = cli.SetFlag(flag, "c", 4).(int)
	// set timeout
	timeout := cli.SetFlag(flag, "t", 2).(int)
	p.timeout = time.Duration(timeout)
	// set method
	p.method = cli.SetFlag(flag, "m", "HEAD").(string)
	// set buff (post)
	buf := cli.SetFlag(flag, "d", "mylg").(string)
	p.buf = buf
	return p, nil
}

// Normalize fixes scheme
func Normalize(URL string) string {
	re := regexp.MustCompile(`(?i)https{0,1}://`)
	if !re.MatchString(URL) {
		URL = fmt.Sprintf("http://%s", URL)
	}
	return URL
}

// pingHeadLoop tries number of connection
// with header information
func (p *Ping) pingHeadLoop() {
	pStrPrefix := "HTTP Response seq=%d, "
	pStrSuffix := "proto=%s, status=%d, time=%.3f ms\n"
	fmt.Printf("HPING %s (%s), Method: HEAD, DNSLookup: %.4f ms\n", p.host, p.rAddr, p.nsTime.Seconds()*1000)
	for i := 0; i < p.count; i++ {
		if r, ok := p.PingHead(); ok {
			fmt.Printf(pStrPrefix+pStrSuffix, i, r.Proto, r.StatusCode, r.TotalTime*1000)
		} else {
			fmt.Printf(pStrPrefix+"timeout\n", i)
		}
	}
}

// pingGetLoop tries number of connection
// with header information
func (p *Ping) pingGetLoop() {
	pStrPrefix := "HTTP Response seq=%d, "
	pStrSuffix := "proto=%s, status=%d, size=%d Bytes, time=%.3f ms\n"
	fmt.Printf("HPING %s (%s), Method: GET, DNSLookup: %.4f ms\n", p.host, p.rAddr, p.nsTime.Seconds()*1000)
	for i := 0; i < p.count; i++ {
		if r, ok := p.PingGet(); ok {
			fmt.Printf(pStrPrefix+pStrSuffix, i, r.Proto, r.StatusCode, r.Size, r.TotalTime*1000)
		} else {
			fmt.Printf(pStrPrefix+"timeout\n", i)
		}
	}
}

// pingGetLoop tries number of connection
// with header information
func (p *Ping) pingPostLoop() {
	pStrPrefix := "HTTP Response seq=%d, "
	pStrSuffix := "proto=%s, status=%d, size=%d Bytes, time=%.3f ms\n"
	fmt.Printf("HPING %s (%s), Method: POST, DNSLookup: %.4f ms\n", p.host, p.rAddr, p.nsTime.Seconds()*1000)
	for i := 0; i < p.count; i++ {
		if r, ok := p.PingPost(); ok {
			fmt.Printf(pStrPrefix+pStrSuffix, i, r.Proto, r.StatusCode, r.Size, r.TotalTime*1000)
		} else {
			fmt.Printf(pStrPrefix+"timeout\n", i)
		}
	}
}

// PingGet tries to ping a web server through http
func (p *Ping) PingGet() (Result, bool) {
	var (
		r     Result
		sTime time.Time
	)

	client := &http.Client{Timeout: p.timeout * time.Second}
	sTime = time.Now()
	resp, err := client.Get(p.url)
	if err != nil {
		return r, false
	}
	defer resp.Body.Close()
	body, err := ioutil.ReadAll(resp.Body)
	r.Size = len(body)
	r.TotalTime = time.Since(sTime).Seconds()
	if err != nil {
		return r, false
	}
	r.StatusCode = resp.StatusCode
	r.Proto = resp.Proto
	return r, true
}

// PingHead tries to ping a web server through http
func (p *Ping) PingHead() (Result, bool) {
	var (
		r     Result
		sTime time.Time
	)

	client := &http.Client{Timeout: p.timeout * time.Second}
	sTime = time.Now()
	resp, err := client.Head(p.url)
	if err != nil {
		return r, false
	}
	r.TotalTime = time.Since(sTime).Seconds()
	if err != nil {
		return r, false
	}
	r.StatusCode = resp.StatusCode
	r.Proto = resp.Proto
	return r, true
}

// PingPost tries to ping a web server through http
func (p *Ping) PingPost() (Result, bool) {
	var (
		r     Result
		sTime time.Time
	)

	client := &http.Client{Timeout: p.timeout * time.Second}
	sTime = time.Now()
	r.Size = len(p.buf)
	reader := strings.NewReader(p.buf)
	resp, err := client.Post(p.url, "text/plain", reader)
	if err != nil {
		return r, false
	}
	r.TotalTime = time.Since(sTime).Seconds()
	if err != nil {
		return r, false
	}
	r.StatusCode = resp.StatusCode
	r.Proto = resp.Proto
	return r, true
}

// Run tries to run ping loop based on the method
func (p *Ping) Run() {
	switch p.method {
	case "HEAD":
		p.pingHeadLoop()
	case "GET":
		p.pingGetLoop()
	case "POST":
		p.pingPostLoop()
	}
}

// help shows ping help
func help() {
	fmt.Println(`
    usage:
          hping [-c count][-t timeout][-m method][-d data] url

    options:		  
          -c count       Send 'count' requests (default: 4)
          -t timeout     Specifies a time limit for requests in second (default is 2) 
          -m method      HTTP methods: GET/POST/HEAD (default: HEAD)
          -d data        Sending the given data (text/json) (default: "mylg")
	`)
}
