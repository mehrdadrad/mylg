package telia

import (
	"errors"
	"io/ioutil"
	"net/http"
	"net/url"
	"regexp"
)

type Provider struct {
	Host string
	IPv  string
	Node string
}

var (
	nodes        = map[string]string{"Amsterdam": "Amsterdam", "Los Angeles": "Los Angeles"}
	dNode string = "Los Angeles"
)

func (p *Provider) Set(host, version string) {
	p.Host = host
	p.IPv = version
	if p.Node == "" {
		p.Node = dNode
	}
}
func (p *Provider) GetDefaultNode() string {
	return dNode
}
func (p *Provider) GetNodes() map[string]string {
	return nodes
}
func (p *Provider) ChangeNode(node string) {
	p.Node = node
}
func (p *Provider) Ping() (string, error) {
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
