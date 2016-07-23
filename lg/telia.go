// Package lg provides looking glass methods for selected looking glasses
// Telia Carrier Looking Glass ASN 1299
package lg

import (
	"bufio"
	"errors"
	"fmt"
	"io/ioutil"
	"net/http"
	"net/url"
	"regexp"
	"sort"
)

// A Telia represents a telia looking glass request
type Telia struct {
	Host  string
	IPv   string
	Node  string
	Nodes []string
}

var teliaDefaultNode = "Los Angeles"

// Set configures host and ip version
func (p *Telia) Set(host, version string) {
	p.Host = host
	p.IPv = version
	if p.Node == "" {
		p.Node = teliaDefaultNode
	}
}

// GetDefaultNode returns telia default node
func (p *Telia) GetDefaultNode() string {
	return teliaDefaultNode
}

// GetNodes returns all Telia nodes (US and International)
func (p *Telia) GetNodes() []string {
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
func (p *Telia) ChangeNode(node string) bool {
	// Validate
	for _, n := range p.Nodes {
		if node == n {
			p.Node = node
			return true
		}
	}
	return false
}

// Ping tries to connect Telia's ping looking glass through HTTP
// Returns the result
func (p *Telia) Ping() (string, error) {
	// Basic validate
	if p.Node == "NA" || len(p.Host) < 5 {
		print("Invalid node or host/ip address")
		return "", errors.New("error")
	}
	resp, err := http.PostForm("http://looking-glass.telia.net/",
		url.Values{"query": {"ping"}, "protocol": {p.IPv}, "addr": {p.Host}, "router": {p.Node}})
	if err != nil {
		return "", err
	}
	if resp.StatusCode != 200 {
		return "", errors.New("error: level3 looking glass is not available")
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

// Trace gets traceroute information from Telia
func (p *Telia) Trace() chan string {
	c := make(chan string)
	resp, err := http.PostForm("http://looking-glass.telia.net/",
		url.Values{"query": {"trace"}, "protocol": {p.IPv}, "addr": {p.Host}, "router": {p.Node}})
	if err != nil {
		println(err)
	}
	go func() {
		defer resp.Body.Close()
		scanner := bufio.NewScanner(resp.Body)
		for scanner.Scan() {
			l := scanner.Text()
			m, _ := regexp.MatchString(`^(traceroute|\s*\d{1,2})`, l)
			if m {
				l = replaceASNTrace(l)
				c <- l
			}
		}
		close(c)
	}()
	return c
}

// BGP gets bgp information from Telia
func (p *Telia) BGP() chan string {
	c := make(chan string)
	resp, err := http.PostForm("http://looking-glass.telia.net/",
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
			if m, _ := regexp.MatchString("Telia Carrier", l); !parse && m {
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
func (p *Telia) FetchNodes() map[string]string {
	var nodes = make(map[string]string, 100)
	resp, err := http.Get("http://looking-glass.telia.net/")
	if err != nil {
		println("error: telia looking glass unreachable (1) ")
		return map[string]string{}
	}
	defer resp.Body.Close()
	body, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		println("error: telia looking glass unreachable (2)" + err.Error())
		return map[string]string{}
	}
	r, _ := regexp.Compile(`(?i)<option value="(?s)([\w|\s|)(._-]+)"> (?s)([\w|\s|)(._-]+)`)
	b := r.FindAllStringSubmatch(string(body), -1)
	for _, v := range b {
		nodes[v[1]] = v[2]
	}
	return nodes
}

//[GOOGLE (ARIN)" HREF="http://www.arin.net/cgi-bin/whois.pl?queryinput=15169" TARGET=_lookup>15169</A>]  1.261 ms 72.14.236.69 (72.14.236.69) [AS  <A title="GOOGLE (ARIN)" HREF="http://www.arin.net/cgi-bin/whois.pl?queryinput=15169" TARGET=_lookup>15169</A>]
// replaceASNTrace
func replaceASNTrace(l string) string {
	m, _ := regexp.MatchString(`\[AS\s+`, l)
	if !m {
		return l
	}
	r := regexp.MustCompile(`(?i)\[AS\s+<A\s+title="([a-z|\d|\s|\(\)_,-]+)"\s+HREF="[a-z|\/|:.-]+\?\w+=\d+"\s+\w+=_lookup>(\d+)</A>]`)
	asn := r.FindStringSubmatch(l)
	if len(asn) == 3 {
		l = r.ReplaceAllString(l, fmt.Sprintf("[%s (%s)]", asn[1], asn[2]))
	}
	return l
}
