package lg

import (
	"errors"
	"io/ioutil"
	"net/http"
	"net/url"
	"regexp"
	"strings"
)

type Level3 struct {
	Host  string
	IPv   string
	Count string
	Node  string
}

var (
	level3Nodes              = map[string]string{"Los Angeles": "ear1.lax1"}
	level3DefaultNode string = "Los Angeles"
)

func sanitize(b string) {

}
func (p *Level3) Set(host, version string) {
	p.Host = host
	p.IPv = version
	p.Count = "5"
	if p.Node == "" {
		p.Node = level3DefaultNode
	}
}
func (p *Level3) GetDefaultNode() string {
	return level3DefaultNode
}
func (p *Level3) GetNodes() map[string]string {
	return level3Nodes
}
func (p *Level3) ChangeNode(node string) {
	p.Node = node
}
func (p *Level3) Ping() (string, error) {
	resp, err := http.PostForm("http://lookingglass.level3.net/ping/lg_ping_output.php",
		url.Values{"count": {p.Count}, "size": {"64"}, "address": {p.Host}, "sitename": {level3Nodes[p.Node]}})
	if err != nil {
		println(err)
	}
	defer resp.Body.Close()
	body, err := ioutil.ReadAll(resp.Body)
	r, _ := regexp.Compile(`</div></div>(?s)(.*?)</font></pre>`)
	b := r.FindStringSubmatch(strings.Replace(string(body), "<br>", "\n", -1))
	if len(b) > 0 {
		return b[1], nil
	} else {
		return "", errors.New("error")
	}
}
