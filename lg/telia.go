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
	teliaNodes              = map[string]string{"Amsterdam": "Amsterdam", "Los Angeles": "Los Angeles"}
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
func (p *Telia) GetNodes() map[string]string {
	return teliaNodes
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
