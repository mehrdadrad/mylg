// Package lg provides looking glass methods for selected looking glasses
// Telia Carrier Looking Glass ASN 1299
package lg

import (
	"errors"
	"io/ioutil"
	"net/http"
	"net/url"
	"regexp"
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
	p.Nodes = nodes
	return nodes
}

// ChangeNode set new requested node
func (p *Telia) ChangeNode(node string) {
	var valid = false
	// Validate
	for _, n := range p.Nodes {
		if node == n {
			valid = true
			break
		}
	}
	if valid {
		p.Node = node
	} else {
		p.Node = "NA"
		println("Invalid node please press tab after node command to show the valid nodes")
	}
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
		println(err)
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
