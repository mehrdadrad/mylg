package lg

import (
	"errors"
	"io/ioutil"
	"net/http"
	"net/url"
	"regexp"
)

type Telia struct {
	Host string
	IPv  string
	Node string
}

var (
	teliaDefaultNode string = "Los Angeles"
)

func (p *Telia) Set(host, version string) {
	p.Host = host
	p.IPv = version
	if p.Node == "" {
		p.Node = teliaDefaultNode
	}
}
func (p *Telia) GetDefaultNode() string {
	return teliaDefaultNode
}
func (p *Telia) GetNodes() []string {
	var nodes []string
	for node := range p.FetchNodes() {
		nodes = append(nodes, node)
	}
	return nodes
}
func (p *Telia) ChangeNode(node string) {
	p.Node = node
}
func (p *Telia) Ping() (string, error) {
	resp, err := http.PostForm("http://looking-glass.telia.net/",
		url.Values{"query": {"ping"}, "protocol": {p.IPv}, "addr": {p.Host}, "router": {p.Node}})
	if err != nil {
		println(err)
	}
	defer resp.Body.Close()
	body, err := ioutil.ReadAll(resp.Body)
	r, _ := regexp.Compile(`<CODE>(?s)(.*?)</CODE>`)
	b := r.FindStringSubmatch(string(body))
	if len(b) > 0 {
		return b[1], nil
	} else {
		return "", errors.New("error")
	}
}
func (p *Telia) FetchNodes() map[string]string {
	var nodes = make(map[string]string, 100)
	resp, err := http.Get("http://looking-glass.telia.net/")
	if err != nil {
		println(err)
		return map[string]string{}
	}
	defer resp.Body.Close()
	body, err := ioutil.ReadAll(resp.Body)
	r, _ := regexp.Compile(`(?i)<option value="(?s)([\w|\s|)(._-]+)"> (?s)([\w|\s|)(._-]+)`)
	b := r.FindAllStringSubmatch(string(body), -1)
	for _, v := range b {
		nodes[v[1]] = v[2]
	}
	return nodes
}
