// Package lg provides looking glass methods for selected looking glasses
// NTT communications Looking Glass
package lg

import (
	"bufio"
	"errors"
	"io/ioutil"
	"net"
	"net/http"
	"net/url"
	"os"
	"os/signal"
	"regexp"
	"sort"
)

type NTT struct {
	Host  string
	IPv   string
	Node  string
	Nodes []string
}

var NTTDefaultNode = "Los Angeles, CA - US"

// Set configures host and ip version
func (p *NTT) Set(host, version string) {
	p.Host = host
	p.IPv = version
	if p.Node == "" {
		p.Node = NTTDefaultNode
	}
}

// GetNodes returns all NTT nodes (US and International)
func (p *NTT) GetNodes() []string {
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

//FetchNodes returns all available nodes through HTTP
func (p *NTT) FetchNodes() map[string]string {
	var nodes = make(map[string]string, 100)
	resp, err := http.Get("http://ssp.pme.gin.ntt.net/lg/lg.cgi")
	if err != nil {
		println("error: NTT looking glass unreachable (1) ")
		return map[string]string{}
	}
	defer resp.Body.Close()
	body, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		println("error: NTT looking glass unreachable (2)" + err.Error())
		return map[string]string{}
	}
	r, _ := regexp.Compile(`(?i)<option value="(?s)([\w|\s|)(,._-]+)"> (?s)([\w|\s|)(,._-]+)`)
	b := r.FindAllStringSubmatch(string(body), -1)
	for _, v := range b {
		nodes[v[1]] = v[2]
	}
	return nodes
}

// ChangeNode set new requested node
func (p *NTT) ChangeNode(node string) bool {
	// Validate
	for _, n := range p.Nodes {
		if node == n {
			p.Node = node
			return true
		}
	}
	return false
}

// GetDefaultNode returns NTT default node
func (p *NTT) GetDefaultNode() string {
	println("please check NTT Com LG terms of use first at https://www.us.ntt.net/support/looking-glass/")
	return NTTDefaultNode
}

// Ping tries to connect NTT's ping looking glass through HTTP
// Returns the result
func (p *NTT) Ping() (string, error) {
	// Basic validate
	if p.Node == "NA" || len(p.Host) < 5 {
		print("Invalid node or host/ip address")
		return "", errors.New("error")
	}
	resp, err := http.PostForm("https://ssp.pme.gin.ntt.net/lg/lg.cgi",
		url.Values{"query": {"ping"}, "protocol": {p.IPv}, "addrFQDN": {p.Host}, "router": {p.Node}, "sourceIP": {"FQDN"}})
	if err != nil {
		return "", err
	}
	if resp.StatusCode != 200 {
		return "", errors.New("error: NTT looking glass is not available")
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
	println(string(body))
	return "", errors.New("error")
}

// Trace gets traceroute information from NTT
func (p *NTT) Trace() chan string {
	c := make(chan string)
	sigCh := make(chan os.Signal, 1)

	signal.Notify(sigCh, os.Interrupt)

	resp, err := http.PostForm("https://ssp.pme.gin.ntt.net/lg/lg.cgi",
		url.Values{"query": {"trace"}, "protocol": {p.IPv}, "addrFQDN": {p.Host}, "router": {p.Node}, "sourceIP": {"FQDN"}})
	if err != nil {
		println(err)
	}
	go func() {
		defer resp.Body.Close()
		scanner := bufio.NewScanner(resp.Body)
	LOOP:
		for scanner.Scan() {
			l := scanner.Text()
			m, _ := regexp.MatchString(`(?i)^(tracing|traceroute|\s*\d{1,2})`, l)
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

// BGP gets bgp information from NTT
func (p *NTT) BGP() chan string {
	c := make(chan string)

	_, _, err := net.ParseCIDR(p.Host)
	if err == nil {
		println("Only IP addresses are allowed for NTT Looking Glass BGP Queries")
	}

	resp, err := http.PostForm("https://ssp.pme.gin.ntt.net/lg/lg.cgi",
		url.Values{"query": {"bgp"}, "protocol": {p.IPv}, "addr": {p.Host}, "router": {p.Node}, "sourceIP": {"IP"}})
	if err != nil {
		println(err)
	}
	// IP addresses are allowed parameters for BGP Queries
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
			if m, _ := regexp.MatchString("Query Results", l); !parse && m {
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
