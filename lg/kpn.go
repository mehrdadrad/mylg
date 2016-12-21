// Package lg provides looking glass methods for selected looking glasses
// KPN Looking Glass ASN 1299
package lg

import (
	"bufio"
	"errors"
	"io/ioutil"
	"net/http"
	"net/url"
	"os"
	"os/signal"
	"regexp"
	"sort"
)

// A KPN represents a KPN looking glass request
type KPN struct {
	Host  string
	IPv   string
	Node  string
	Nodes []string
}

var KPNDefaultNode = "Amsterdam (NL)"

// Set configures host and ip version
func (p *KPN) Set(host, version string) {
	p.Host = host
	p.IPv = version
	if p.Node == "" {
		p.Node = KPNDefaultNode
	}
}

// GetDefaultNode returns KPN default node
func (p *KPN) GetDefaultNode() string {
	return KPNDefaultNode
}

// GetNodes returns all KPN nodes (US and International)
func (p *KPN) GetNodes() []string {
	// Memory cache
	if len(p.Nodes) > 1 {
		return p.Nodes
	}
	var nodes []string
	for node := range p.FetchNodes() {
		nodes = append(nodes, node)
	}
	sort.Strings(nodes)
	p.Nodes = nodes
	return nodes
}

// ChangeNode set new requested node
func (p *KPN) ChangeNode(node string) bool {
	// Validate
	for _, n := range p.Nodes {
		if node == n {
			p.Node = node
			return true
		}
	}
	return false
}

// Ping tries to connect KPN's ping looking glass through HTTP
// Returns the result
func (p *KPN) Ping() (string, error) {
	// Basic validate
	if p.Node == "NA" || len(p.Host) < 5 {
		print("Invalid node or host/ip address")
		return "", errors.New("error")
	}
	resp, err := http.PostForm("http://lg.eurorings.net/index.cgi",
		url.Values{"query": {"ping"}, "protocol": {p.IPv}, "addr": {p.Host}, "router": {p.Node}})
	if err != nil {
		return "", err
	}
	if resp.StatusCode != 200 {
		return "", errors.New("error: KPN looking glass is not available")
	}
	defer resp.Body.Close()
	body, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		return "", err
	}
	r, _ := regexp.Compile(`<CODE>(?s)(.*?)</CODE>`)
	b := r.FindStringSubmatch(string(body))
	if len(b) > 0 {
		return b[1], nil
	}
	return "", errors.New("error")
}

// Trace gets traceroute information from KPN
func (p *KPN) Trace() chan string {
	c := make(chan string)
	sigCh := make(chan os.Signal, 1)

	signal.Notify(sigCh, os.Interrupt)

	resp, err := http.PostForm("http://lg.eurorings.net/index.cgi",
		url.Values{"query": {"trace"}, "protocol": {p.IPv}, "addr": {p.Host}, "router": {p.Node}})
	if err != nil {
		println(err)
	}
	go func() {
		defer resp.Body.Close()
		scanner := bufio.NewScanner(resp.Body)
	LOOP:
		for scanner.Scan() {
			l := scanner.Text()
			m, _ := regexp.MatchString(`^(traceroute|\s*\d{1,2})`, l)
			if m {
				l = replaceASNTrace(l)
				select {
				case <-sigCh:
					break LOOP
				case c <- l:
				}
			}
		}
		signal.Stop(sigCh)
		close(c)
	}()
	return c
}

// BGP gets bgp information from KPN
func (p *KPN) BGP() chan string {
	c := make(chan string)
	resp, err := http.PostForm("http://lg.eurorings.net/index.cgi",
		url.Values{"query": {"bgp"}, "protocol": {p.IPv}, "addr": {p.Host}, "router": {p.Node}})
	if err != nil {
		println(err)
	}
	go func() {
		var (
			parse = false
			last  string
		)
		defer resp.Body.Close()
		scanner := bufio.NewScanner(resp.Body)
		for scanner.Scan() {
			l := scanner.Text()
			l = sanitize(l)
			if m, _ := regexp.MatchString("Location:", l); !parse && m {
				parse = true
				continue
			}
			if !parse || (l == last) {
				continue
			}
			c <- l
			last = l
		}
		close(c)
	}()
	return c
}

//FetchNodes returns all available nodes through HTTP
func (p *KPN) FetchNodes() map[string]string {
	var nodes = make(map[string]string, 100)
	resp, err := http.Get("http://lg.eurorings.net/index.cgi")
	if err != nil {
		println("error: KPN looking glass unreachable (1) ")
		return map[string]string{}
	}
	defer resp.Body.Close()
	body, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		println("error: KPN looking glass unreachable (2)" + err.Error())
		return map[string]string{}
	}
	r, _ := regexp.Compile(`(?i)<option value="(?s)([\w|\s|)(._-]+)"> (?s)([\w|\s|)(._-]+)`)
	b := r.FindAllStringSubmatch(string(body), -1)
	for _, v := range b {
		nodes[v[1]] = v[2]
	}
	return nodes
}
