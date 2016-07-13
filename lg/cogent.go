// Package lg provides looking glass methods for selected looking glasses
// Cogent Carrier Looking Glass ASN 174
package lg

import (
	"bufio"
	"errors"
	"io/ioutil"
	"net/http"
	"net/url"
	"regexp"
	"sort"
)

// A Cogent represents a telia looking glass request
type Cogent struct {
	Host  string
	IPv   string
	Node  string
	Nodes []string
}

var (
	cogentNodes       = map[string]string{}
	cogentDefaultNode = "US - Los Angeles"
)

// Set configures host and ip version
func (p *Cogent) Set(host, version string) {
	p.Host = host
	p.IPv = version
	if p.Node == "" {
		p.Node = cogentDefaultNode
	}
}

// GetDefaultNode returns telia default node
func (p *Cogent) GetDefaultNode() string {
	return cogentDefaultNode
}

// GetNodes returns all Cogent nodes (US and International)
func (p *Cogent) GetNodes() []string {
	// Memory cache
	if len(p.Nodes) > 1 {
		return p.Nodes
	}
	cogentNodes = p.FetchNodes()
	var nodes []string
	for node := range cogentNodes {
		nodes = append(nodes, node)
	}
	sort.Strings(nodes)
	p.Nodes = nodes
	return nodes
}

// ChangeNode set new requested node
func (p *Cogent) ChangeNode(node string) {
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

// Ping tries to connect Cogent's ping looking glass through HTTP
// Returns the result
func (p *Cogent) Ping() (string, error) {
	// Basic validate
	if p.Node == "NA" || len(p.Host) < 5 {
		print("Invalid node or host/ip address")
		return "", errors.New("error")
	}
	var cmd = "P4"
	if p.IPv == "ipv6" {
		cmd = "P6"
	}
	resp, err := http.PostForm("http://www.cogentco.com/lookingglass.php",
		url.Values{"FKT": {"go!"}, "CMD": {cmd}, "DST": {p.Host}, "LOC": {cogentNodes[p.Node]}})
	if err != nil {
		println(err)
		return "", err
	}
	defer resp.Body.Close()
	body, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		return "", err
	}
	r, _ := regexp.Compile(`<pre>(?s)(.*?)</pre>`)
	b := r.FindStringSubmatch(string(body))
	if len(b) > 0 {
		return b[1], nil
	}
	return "", errors.New("error")
}

// Trace gets traceroute information from Cogent
func (p *Cogent) Trace() chan string {
	c := make(chan string)
	var cmd = "T4"
	if p.IPv == "ipv6" {
		cmd = "T6"
	}
	resp, err := http.PostForm("http://www.cogentco.com/lookingglass.php",
		url.Values{"FKT": {"go!"}, "CMD": {cmd}, "DST": {p.Host}, "LOC": {cogentNodes[p.Node]}})
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

// BGP gets bgp information from cogent
func (p *Cogent) BGP() chan string {
	c := make(chan string)
	resp, err := http.PostForm("http://www.cogentco.com/lookingglass.php",
		url.Values{"FKT": {"go!"}, "CMD": {"BGP"}, "DST": {p.Host}, "LOC": {cogentNodes[p.Node]}})
	if err != nil {
		println(err)
	}
	go func() {
		defer resp.Body.Close()
		scanner := bufio.NewScanner(resp.Body)
		for scanner.Scan() {
			l := scanner.Text()
			c <- l
		}
		close(c)
	}()
	return c
}

//FetchNodes returns all available nodes through HTTP
func (p *Cogent) FetchNodes() map[string]string {
	var nodes = make(map[string]string, 100)
	resp, err := http.Get("http://www.cogentco.com/lookingglass.php")
	if err != nil {
		println("error: cogent looking glass unreachable (1)")
		return map[string]string{}
	}
	defer resp.Body.Close()
	body, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		println("error: cogent looking glass unreachable (2)" + err.Error())
		return map[string]string{}
	}
	r, _ := regexp.Compile(`(?is)Option\(\"([\w|\s|-]+)\",\"([\w|\d]+)\"`)
	b := r.FindAllStringSubmatch(string(body), -1)
	for _, v := range b {
		nodes[v[1]] = v[2]
	}
	return nodes
}
