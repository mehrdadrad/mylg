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
	level3Nodes              = map[string]string{"Los Angeles, CA": "ear1.lax1"}
	level3DefaultNode string = "Los Angeles, CA"
)

func sanitize(b string) string {
	re := regexp.MustCompile("<(.*)>")
	return re.ReplaceAllString(b, "")
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
	level3Nodes = p.FetchNodes()
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
		return sanitize(b[1]), nil
	} else {
		return "", errors.New("error")
	}
}

func (p *Level3) FetchNodes() map[string]string {
	var nodes = make(map[string]string, 100)
	resp, err := http.Get("http://lookingglass.level3.net/ping/lg_ping_main.php")
	if err != nil {
		println(err)
		return map[string]string{}
	}
	defer resp.Body.Close()
	body, err := ioutil.ReadAll(resp.Body)
	r, _ := regexp.Compile(`(?i)<option value="(?s)([\w|\s|)(._-]+)">(?s)([a-z|\s|)(,._-]+)</option>`)
	b := r.FindAllStringSubmatch(string(body), -1)
	for _, v := range b {
		nodes[v[2]] = v[1]
	}
	return nodes
}
