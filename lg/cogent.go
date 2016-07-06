// Package lg provides looking glass methods for selected looking glasses
// Cogent Carrier Looking Glass ASN 174
package lg

import (
	"errors"
	"io/ioutil"
	"net/http"
	"net/url"
	"regexp"
)

// A Cogent represents a telia looking glass request
type Cogent struct {
	Host string
	IPv  string
	Node string
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
	cogentNodes = p.FetchNodes()
	var nodes []string
	for node := range cogentNodes {
		nodes = append(nodes, node)
	}
	return nodes
}

// ChangeNode set new requested node
func (p *Cogent) ChangeNode(node string) {
	p.Node = node
}

// Ping tries to connect Cogent's ping looking glass through HTTP
// Returns the result
func (p *Cogent) Ping() (string, error) {
	var cmd = "P4"
	if p.IPv == "ipv6" {
		cmd = "P6"
	}
	resp, err := http.PostForm("http://www.cogentco.com/lookingglass.php",
		url.Values{"FKT": {"go!"}, "CMD": {cmd}, "DST": {p.Host}, "LOC": {cogentNodes[p.Node]}})
	if err != nil {
		println(err)
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

//FetchNodes returns all available nodes through HTTP
func (p *Cogent) FetchNodes() map[string]string {
	var nodes = make(map[string]string, 100)
	resp, err := http.Get("http://www.cogentco.com/lookingglass.php")
	if err != nil {
		println("error: cogent looking glass unreachable (1)" + err.Error())
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
