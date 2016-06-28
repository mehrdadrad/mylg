package telia

import (
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

func Init(host, version, node string) *Telia {
	return &Telia{host, version, node}
}

func (t *Telia) Ping() string {
	resp, err := http.PostForm("http://looking-glass.telia.net/",
		url.Values{"query": {"ping"}, "protocol": {t.IPv}, "addr": {t.Host}, "router": {t.Node}})
	if err != nil {
		println(err)
	}
	defer resp.Body.Close()
	body, err := ioutil.ReadAll(resp.Body)
	r, _ := regexp.Compile(`<CODE>(?s)(.*?)</CODE>`)
	p := r.FindStringSubmatch(string(body))
	return p[1]
}
