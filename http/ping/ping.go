// Package ping tries to ping a HTTP server through different ways
// Connection, Session (Head), Get and Post
package ping

import (
	"fmt"
	"io/ioutil"
	"net"
	"net/http"
	"net/url"
	"os"
	"os/signal"
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
func NewPing(args string, cfg cli.Config) (*Ping, error) {
	URL, flag := cli.Flag(args)
	// help
	if _, ok := flag["help"]; ok || URL == "" {
		help(cfg)
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
	p.count = cli.SetFlag(flag, "c", cfg.Hping.Count).(int)
	// set timeout
	timeout := cli.SetFlag(flag, "t", cfg.Hping.Timeout).(string)
	p.timeout, err = time.ParseDuration(timeout)
	if err != nil {
		return p, fmt.Errorf("Failed to parse config.hping.timeout: %s. Correct syntax is <number>s/ms", err)
	}
	// set method
	p.method = cli.SetFlag(flag, "m", cfg.Hping.Method).(string)
	p.method = strings.ToUpper(p.method)
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

// Run tries to ping w/ pretty print
func (p *Ping) Run() {
	if p.method != "GET" && p.method != "POST" && p.method != "HEAD" {
		fmt.Printf("Error: Method '%s' not recognized.\n", p.method)
		return
	}
	var (
		sigCh = make(chan os.Signal, 1)
		c     = make(map[int]float64, 10)
		s     []float64
	)
	// capture interrupt w/ s channel
	signal.Notify(sigCh, os.Interrupt)
	defer signal.Stop(sigCh)

	pStrPrefix := "HTTP Response seq=%d, "
	pStrSuffix := "proto=%s, status=%d, size=%d Bytes, time=%.3f ms\n"
	pStrSuffixHead := "proto=%s, status=%d, time=%.3f ms\n"
	fmt.Printf("HPING %s (%s), Method: %s, DNSLookup: %.4f ms\n", p.host, p.rAddr, p.method, p.nsTime.Seconds()*1000)

LOOP:
	for i := 0; i < p.count; i++ {
		if r, err := p.Ping(); err == nil {
			if p.method != "HEAD" {
				fmt.Printf(pStrPrefix+pStrSuffix, i, r.Proto, r.StatusCode, r.Size, r.TotalTime*1000)
			} else {
				fmt.Printf(pStrPrefix+pStrSuffixHead, i, r.Proto, r.StatusCode, r.TotalTime*1000)
			}
			c[r.StatusCode]++
			s = append(s, r.TotalTime*1000)
		} else {
			c[-1]++
			errmsg := strings.Split(err.Error(), ": ")
			fmt.Printf(pStrPrefix+"%s\n", i, errmsg[len(errmsg)-1])
		}
		select {
		case <-sigCh:
			break LOOP
		default:
		}
	}
	// print statistics
	printStats(c, s, p.host)
}

// printStats prints out the footer
func printStats(c map[int]float64, s []float64, host string) {
	var r = make(map[string]float64, 5)

	// total replied requests
	for k, v := range c {
		if k < 0 {
			continue
		}
		r["sum"] += v
	}

	for _, v := range s {
		// maximum
		if r["max"] < v {
			r["max"] = v
		}
		// minimum
		if r["min"] > v || r["min"] == 0 {
			r["min"] = v
		}
		// average
		if r["avg"] == 0 {
			r["avg"] = v
		} else {
			r["avg"] = (r["avg"] + v) / 2
		}
	}

	totalReq := r["sum"] + c[-1]
	failPct := 100 - (100*r["sum"])/totalReq

	fmt.Printf("\n--- %s HTTP ping statistics --- \n", host)
	fmt.Printf("%.0f requests transmitted, %.0f replies received, %.0f%% requests failed\n", totalReq, r["sum"], failPct)
	fmt.Printf("HTTP Round-trip min/avg/max = %.2f/%.2f/%.2f ms\n", r["min"], r["avg"], r["max"])
	for k, v := range c {
		if k < 0 {
			continue
		}
		progress := fmt.Sprintf("%10s", strings.Repeat("\u2588", int(v*100/(r["sum"])/5)))
		fmt.Printf("HTTP Code [%d] responses : [%s] %.2f%% \n", k, progress, v*100/(r["sum"]))
	}
}

// Ping tries to ping a web server through http
func (p *Ping) Ping() (Result, error) {
	var (
		r     Result
		sTime time.Time
		resp  *http.Response
		req   *http.Request
		err   error
	)

	client := &http.Client{Timeout: p.timeout}
	sTime = time.Now()

	if p.method == "POST" {
		r.Size = len(p.buf)
		reader := strings.NewReader(p.buf)
		req, err = http.NewRequest(p.method, p.url, reader)
	} else {
		req, err = http.NewRequest(p.method, p.url, nil)
	}

	if err != nil {
		return r, err
	}

	req.Header.Add("User-Agent", "myLG (http://mylg.io)")

	resp, err = client.Do(req)

	if err != nil {
		return r, err
	}

	r.TotalTime = time.Since(sTime).Seconds()

	if p.method == "GET" {
		defer resp.Body.Close()
		body, err := ioutil.ReadAll(resp.Body)
		if err != nil {
			return r, err
		}
		r.Size = len(body)
	}

	r.StatusCode = resp.StatusCode
	r.Proto = resp.Proto
	return r, nil
}

// help shows ping help
func help(cfg cli.Config) {
	fmt.Printf(`
    usage:
          hping url [options]

    options:
          -c count       Send 'count' requests (default: %d)
          -t timeout     Specifies a time limit for requests in ms/s (default is %s)
          -m method      HTTP methods: GET/POST/HEAD (default: %s)
          -d data        Sending the given data (text/json) (default: "%s")
	`,
		cfg.Hping.Count,
		cfg.Hping.Timeout,
		cfg.Hping.Method,
		cfg.Hping.Data)
}
