// Package ping tries to ping a HTTP server through different ways
// Connection, Session (Head), Get and Post
package ping

import (
	"crypto/tls"
	"encoding/json"
	"fmt"
	"io"
	"io/ioutil"
	"net"
	"net/http"
	"net/http/httptrace"
	"net/url"
	"os"
	"os/signal"
	"regexp"
	"strings"
	"time"

	"github.com/mehrdadrad/mylg/cli"
)

var stdout *os.File

// Ping represents HTTP ping request
type Ping struct {
	url           string
	host          string
	interval      time.Duration
	timeout       time.Duration
	count         int
	method        string
	uAgent        string
	proxy         *url.URL
	buf           string
	rAddr         net.Addr
	nsTime        time.Duration
	conn          net.Conn
	quiet         bool
	dCompress     bool
	kAlive        bool
	TLSSkipVerify bool
	tracerEnabled bool
	fmtJSON       bool
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
	Trace      Trace
}

// Trace holds trace results
type Trace struct {
	ConnectionTime  float64
	TimeToFirstByte float64
}

// PrintPingResult prints result from each individual ping
func (r Result) PrintPingResult(p *Ping, seq int, err error) {
	pStrPrefix := "HTTP Response seq=%d, "
	pStrSuffix := "proto=%s, status=%d, size=%d Bytes, time=%.3f ms"
	pStrSuffixHead := "proto=%s, status=%d, time=%.3f ms"
	pStrTrace := ", connection=%.3f ms, first byte read=%.3f ms\n"

	if p.quiet {
		if err != nil {
			fmt.Printf("!")
			return
		}
		fmt.Printf(".")
		return
	}

	if err != nil {
		errmsg := strings.Split(err.Error(), ": ")
		fmt.Printf(pStrPrefix+"%s\n", seq, errmsg[len(errmsg)-1])
		return
	}

	if p.method == "HEAD" {
		if p.tracerEnabled {
			fmt.Printf(pStrPrefix+pStrSuffixHead+pStrTrace, seq, r.Proto, r.StatusCode, r.TotalTime*1e3, r.Trace.ConnectionTime, r.Trace.TimeToFirstByte)
			return
		}
		fmt.Printf(pStrPrefix+pStrSuffixHead+"\n", seq, r.Proto, r.StatusCode, r.TotalTime*1e3)
		return
	}
	if p.tracerEnabled {
		fmt.Printf(pStrPrefix+pStrSuffix+pStrTrace, seq, r.Proto, r.StatusCode, r.Size, r.TotalTime*1e3, r.Trace.ConnectionTime, r.Trace.TimeToFirstByte)
		return
	}
	fmt.Printf(pStrPrefix+pStrSuffix+"\n", seq, r.Proto, r.StatusCode, r.Size, r.TotalTime*1e3)
	return
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
		url:           URL,
		host:          u.Host,
		rAddr:         ipAddr,
		count:         cli.SetFlag(flag, "c", cfg.Hping.Count).(int),
		tracerEnabled: cli.SetFlag(flag, "trace", false).(bool),
		fmtJSON:       cli.SetFlag(flag, "json", false).(bool),
		uAgent:        cli.SetFlag(flag, "u", "myLG (http://mylg.io)").(string),
		dCompress:     cli.SetFlag(flag, "dc", false).(bool),
		kAlive:        cli.SetFlag(flag, "k", false).(bool),
		TLSSkipVerify: cli.SetFlag(flag, "nc", false).(bool),
		quiet:         cli.SetFlag(flag, "q", false).(bool),
		nsTime:        time.Since(sTime),
	}

	// set interval
	interval := cli.SetFlag(flag, "i", "0s").(string)
	p.interval, err = time.ParseDuration(interval)
	if err != nil {
		return p, fmt.Errorf("Failed to parse interval: %s. Correct syntax is <number>s/ms", err)
	}
	// set timeout
	timeout := cli.SetFlag(flag, "t", cfg.Hping.Timeout).(string)
	p.timeout, err = time.ParseDuration(timeout)
	if err != nil {
		return p, fmt.Errorf("Failed to parse timeout: %s. Correct syntax is <number>s/ms", err)
	}
	// set method
	p.method = cli.SetFlag(flag, "m", cfg.Hping.Method).(string)
	p.method = strings.ToUpper(p.method)
	// set proxy
	proxy := cli.SetFlag(flag, "p", "").(string)
	if pURL, err := url.Parse(proxy); err == nil {
		p.proxy = pURL
	} else {
		return p, fmt.Errorf("Failed to parse proxy url: %v", err)
	}
	// set buff (post)
	buf := cli.SetFlag(flag, "d", "mylg").(string)
	p.buf = buf

	if p.fmtJSON {
		muteStdout()
	}

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

	fmt.Printf("HPING %s (%s), Method: %s, DNSLookup: %.4f ms\n", p.host, p.rAddr, p.method, p.nsTime.Seconds()*1e3)

LOOP:
	for i := 0; i < p.count; i++ {
		if r, err := p.Ping(); err == nil {
			r.PrintPingResult(p, i, err)
			c[r.StatusCode]++
			s = append(s, r.TotalTime*1e3)
		} else {
			c[-1]++
			r.PrintPingResult(p, i, err)
		}
		select {
		case <-sigCh:
			break LOOP
		default:
		}
		time.Sleep(p.interval)
	}

	// print statistics
	if p.fmtJSON {
		unMuteStdout()
		p.printStatsJSON(c, s)
	} else {
		p.printStats(c, s)
	}
}

// printStats prints out the footer
func (p *Ping) printStats(c map[int]float64, s []float64) {

	r := calcStats(c, s)

	totalReq := r["sum"] + c[-1]
	failPct := 100 - (100*r["sum"])/totalReq

	fmt.Printf("\n--- %s HTTP ping statistics --- \n", p.host)
	fmt.Printf("%.0f requests transmitted, %.0f replies received, %.0f%% requests failed\n", totalReq, r["sum"], failPct)
	fmt.Printf("HTTP Round-trip min/avg/max = %.2f/%.2f/%.2f ms\n", r["min"], r["avg"], r["max"])
	for k, v := range c {
		if k < 0 {
			continue
		}
		progress := fmt.Sprintf("%-20s", strings.Repeat("\u2588", int(v*100/(totalReq)/5)))
		fmt.Printf("HTTP Code [%d] responses : [%s] %.2f%% \n", k, progress, v*100/(totalReq))
	}
}

// printStats prints out in json format
func (p *Ping) printStatsJSON(c map[int]float64, s []float64) {
	var statusCode = make(map[int]float64, 10)

	r := calcStats(c, s)

	totalReq := r["sum"] + c[-1]
	failPct := 100 - (100*r["sum"])/totalReq

	for k, v := range c {
		if k < 0 {
			continue
		}
		statusCode[k] = v * 100 / (totalReq)
	}

	trace := struct {
		Host      string  `json:"host"`
		DNSLookup float64 `json:"dnslookup"`
		Count     int     `json:"count"`

		Min float64 `json:"min"`
		Avg float64 `json:"avg"`
		Max float64 `json:"max"`

		Failure     float64         `json:"failure"`
		StatusCodes map[int]float64 `json:"statuscodes"`
	}{
		p.host,
		p.nsTime.Seconds() * 1e3,
		p.count,

		r["min"],
		r["avg"],
		r["max"],

		failPct,
		statusCode,
	}

	b, err := json.Marshal(trace)
	if err != nil {

	}

	fmt.Println(string(b))
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

	tr := &http.Transport{
		DisableKeepAlives:  !p.kAlive,
		DisableCompression: p.dCompress,
		TLSClientConfig: &tls.Config{
			InsecureSkipVerify: p.TLSSkipVerify,
		},
	}

	if p.proxy.String() != "" {
		tr.Proxy = http.ProxyURL(p.proxy)
	}

	client := &http.Client{
		CheckRedirect: func(req *http.Request, via []*http.Request) error {
			// Don't follow redirects
			return http.ErrUseLastResponse
		},
		Timeout:   p.timeout,
		Transport: tr,
	}

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

	// customized header
	req.Header.Add("User-Agent", p.uAgent)
	// context, tracert
	if p.tracerEnabled && !p.quiet {
		req = req.WithContext(httptrace.WithClientTrace(req.Context(), tracer(&r)))
	}
	resp, err = client.Do(req)

	if err != nil {
		return r, err
	}
	defer resp.Body.Close()

	r.TotalTime = time.Since(sTime).Seconds()

	if p.method == "GET" {
		body, err := ioutil.ReadAll(resp.Body)
		if err != nil {
			return r, err
		}
		r.Size = len(body)
	} else {
		io.Copy(ioutil.Discard, resp.Body)
	}

	r.StatusCode = resp.StatusCode
	r.Proto = resp.Proto
	return r, nil
}

func tracer(r *Result) *httptrace.ClientTrace {
	var (
		begin   = time.Now()
		elapsed time.Duration
	)

	return &httptrace.ClientTrace{
		ConnectDone: func(network, addr string, err error) {
			elapsed = time.Since(begin)
			begin = time.Now()
			r.Trace.ConnectionTime = elapsed.Seconds() * 1e3
		},
		GotFirstResponseByte: func() {
			elapsed = time.Since(begin)
			begin = time.Now()
			r.Trace.TimeToFirstByte = elapsed.Seconds() * 1e3
		},
	}
}

func calcStats(c map[int]float64, s []float64) map[string]float64 {
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
	return r
}

func muteStdout() {
	stdout = os.Stdout
	_, w, _ := os.Pipe()
	os.Stdout = w
}

func unMuteStdout() {
	os.Stdout = stdout
}

// help shows ping help
func help(cfg cli.Config) {
	fmt.Printf(`
    usage:
          hping url [options]

    options:
          -c   count        Send 'count' requests (default: %d)
          -t   timeout      Set a time limit for requests in ms/s (default is %s)
          -i   interval     Set a wait time between sending each request in ms/s
          -m   method       HTTP methods: GET/POST/HEAD (default: %s)
          -d   data         Sending the given data (text/json) (default: "%s")
          -p   proxy server Set proxy http://url:port
          -u   user agent   Set user agent
          -q                Quiet reqular output
          -k                Enable keep alive
          -dc               Disable compression
          -nc               Don’t check the server certificate
          -trace            Provides the events within client requests
          -json             Export statistics as json format
		  `,
		cfg.Hping.Count,
		cfg.Hping.Timeout,
		cfg.Hping.Method,
		cfg.Hping.Data)
}
