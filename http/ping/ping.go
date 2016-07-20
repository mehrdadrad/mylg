// Package ping tries to ping a HTTP server through different ways
// Connection, Session (Head), Get and Post
package ping

import (
	"bufio"
	"fmt"
	"io/ioutil"
	"net"
	"net/http"
	"net/url"
	"reflect"
	"regexp"
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
func NewPing(URL string, timeout time.Duration) (*Ping, error) {
	URL, flag := cli.Flag(URL)
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
		url:     URL,
		host:    u.Host,
		rAddr:   ipAddr,
		nsTime:  time.Since(sTime),
		timeout: timeout,
	}
	// set count
	p.count = cli.SetFlag(flag, "c", 4).(int)
	// set method
	p.method = cli.SetFlag(flag, "m", "HEAD").(string)

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

// pingHeadLoop tries number of connection
// with header information
func (p *Ping) pingHeadLoop() {
	pStrPrefix := "HTTP Response seq=%d, "
	pStrPostfix := "proto=%s, status=%d, time=%.3f ms\n"
	fmt.Printf("HPING %s (%s), Method: HEAD, DNSLookup: %.4f ms\n", p.host, p.rAddr, p.nsTime.Seconds()*1000)
	for i := 0; i < p.count; i++ {
		if r, ok := p.pingHead(); ok {
			fmt.Printf(pStrPrefix+pStrPostfix, i, r.Proto, r.StatusCode, r.TotalTime*1000)
		} else {
			fmt.Printf(pStrPrefix+"timeout\n", i)
		}
	}
}

// pingHeadLoop tries number of connection
// with header information
func (p *Ping) pingGetLoop() {
	pStrPrefix := "HTTP Response seq=%d, "
	pStrPostfix := "proto=%s, status=%d, size=%d Bytes, time=%.3f ms\n"
	fmt.Printf("HPING %s (%s), Method: GET, DNSLookup: %.4f ms\n", p.host, p.rAddr, p.nsTime.Seconds()*1000)
	for i := 0; i < p.count; i++ {
		if r, ok := p.pingGet(); ok {
			fmt.Printf(pStrPrefix+pStrPostfix, i, r.Proto, r.StatusCode, r.Size, r.TotalTime*1000)
		} else {
			fmt.Printf(pStrPrefix+"timeout\n", i)
		}
	}
}

// pingGet tries to ping a web server through http
func (p *Ping) pingGet() (Result, bool) {
	var (
		r     Result
		sTime time.Time
	)

	client := &http.Client{Timeout: 2 * time.Second}
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

// pingHead tries to ping a web server through http
func (p *Ping) pingHead() (Result, bool) {
	var (
		r     Result
		sTime time.Time
	)

	client := &http.Client{Timeout: 2 * time.Second}
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

// pingNetHead tries to execute head command
func (p *Ping) pingNetHead() (Result, bool) {
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

// Run tries to run ping loop based on the method
func (p *Ping) Run() {
	switch p.method {
	case "HEAD":
		p.pingHeadLoop()
	case "GET":
		p.pingGetLoop()
	}
}

// help shows ping help
func help() {
	println(`
    usage:
          hping [-c count][-m method] url
	`)
}
