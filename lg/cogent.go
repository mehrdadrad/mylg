package lg

import (
	"errors"
	"io/ioutil"
	"net/http"
	"net/url"
	"regexp"
)

type Cogent struct {
	Host string
	IPv  string
	Node string
}

var (
	cogentNodes              = map[string]string{}
	cogentDefaultNode string = "US - Los Angeles"
)

func (p *Cogent) Set(host, version string) {
	p.Host = host
	p.IPv = version
	if p.Node == "" {
		p.Node = cogentDefaultNode
	}
}
func (p *Cogent) GetDefaultNode() string {
	return cogentDefaultNode
}
func (p *Cogent) GetNodes() []string {
	cogentNodes = p.FetchNodes()
	var nodes []string
	for node := range cogentNodes {
		nodes = append(nodes, node)
	}
	return nodes
}
func (p *Cogent) ChangeNode(node string) {
	p.Node = node
}
func (p *Cogent) Ping() (string, error) {
	var cmd string = "P4"
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
	r, _ := regexp.Compile(`<pre>(?s)(.*?)</pre>`)
	b := r.FindStringSubmatch(string(body))
	if len(b) > 0 {
		return b[1], nil
	} else {
		return "", errors.New("error")
	}
}
func (p *Cogent) FetchNodes() map[string]string {
	var nodes = make(map[string]string, 100)
	resp, err := http.Get("http://www.cogentco.com/lookingglass.php")
	if err != nil {
		println(err)
		return map[string]string{}
	}
	defer resp.Body.Close()
	body, err := ioutil.ReadAll(resp.Body)
	r, _ := regexp.Compile(`(?is)Option\(\"([\w|\s|-]+)\",\"([\w|\d]+)\"`)
	b := r.FindAllStringSubmatch(string(body), -1)
	for _, v := range b {
		nodes[v[1]] = v[2]
	}
	return nodes
}
