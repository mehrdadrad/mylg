// Package lg provides looking glass methods for selected looking glasses
// Level3 Carrier Looking Glass ASN 3356
package lg

import (
	"bufio"
	"errors"
	"html"
	"io/ioutil"
	"net/http"
	"net/url"
	"regexp"
	"sort"
	"strings"
)

// A Level3 represents a telia looking glass request
type Level3 struct {
	Host  string
	CIDR  string
	IPv   string
	Count string
	Node  string
	Nodes []string
}

var (
	level3Nodes       = map[string]string{"Los Angeles, CA": "ear1.lax1"}
	level3DefaultNode = "Los Angeles, CA"
)

// sanitize removes html tags
func sanitize(b string) string {
	re := regexp.MustCompile(`<br>`)
	b = re.ReplaceAllString(b, "\n")
	re = regexp.MustCompile(`<[^>]*>`)
	b = re.ReplaceAllString(b, "")
	return html.UnescapeString(b)
}

// Set configures host and ip version
func (p *Level3) Set(host, version string) {
	if i := strings.Index(host, "/"); i > 0 {
		p.Host = host[:i]
		p.CIDR = host[i+1:]
	} else {
		p.Host = host
		p.CIDR = "24"
	}
	p.IPv = version
	p.Count = "5"
	if p.Node == "" {
		p.Node = level3DefaultNode
	}
}

// GetDefaultNode returns telia default node
func (p *Level3) GetDefaultNode() string {
	return level3DefaultNode
}

// GetNodes returns all level3 nodes (US and International)
func (p *Level3) GetNodes() []string {
	// Memory cache
	if len(p.Nodes) > 1 {
		return p.Nodes
	}
	level3Nodes = p.FetchNodes()
	var nodes []string
	for node := range level3Nodes {
		nodes = append(nodes, node)
	}
	sort.Strings(nodes)
	p.Nodes = nodes
	return nodes
}

// ChangeNode set new requested node
func (p *Level3) ChangeNode(node string) {
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

// Ping tries to connect Level3's ping looking glass through HTTP
// Returns the result
func (p *Level3) Ping() (string, error) {
	// Basic validate
	if p.Node == "NA" || len(p.Host) < 5 {
		print("Invalid node or host/ip address")
		return "", errors.New("error")
	}
	resp, err := http.PostForm("http://lookingglass.level3.net/ping/lg_ping_output.php",
		url.Values{"count": {p.Count}, "size": {"64"}, "address": {p.Host}, "sitename": {level3Nodes[p.Node]}})
	if err != nil {
		println(err)
		return "", err
	}
	defer resp.Body.Close()
	body, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		return "", err
	}
	r, _ := regexp.Compile(`</div></div>(?s)(.*?)</font></pre>`)
	b := r.FindStringSubmatch(strings.Replace(string(body), "<br>", "\n", -1))
	if len(b) > 0 {
		return sanitize(b[1]), nil
	}
	return "", errors.New("error")
}

// Trace gets traceroute information from level3
func (p *Level3) Trace() chan string {
	c := make(chan string)
	resp, err := http.PostForm("http://lookingglass.level3.net/traceroute/lg_tr_output.php",
		url.Values{"address": {p.Host}, "sitename": {level3Nodes[p.Node]}})
	if err != nil {
		println(err)
	}
	go func() {
		defer resp.Body.Close()
		scanner := bufio.NewScanner(resp.Body)
		for scanner.Scan() {
			l := scanner.Text()
			m, _ := regexp.MatchString(`(?i)(^traceroute|\s+\d{1,2})\s+`, l)
			if m {
				l = sanitize(l)
				c <- l
			}
		}
		close(c)
	}()
	return c
}

// BGP gets bgp information
func (p *Level3) BGP() chan string {
	c := make(chan string)
	resp, err := http.PostForm("http://lookingglass.level3.net/bgp/lg_bgp_output.php",
		url.Values{"address": {p.Host}, "length": {p.CIDR}, "sitename": {level3Nodes[p.Node]}})
	if err != nil {
		println(err.Error())
		close(c)
	}
	go func() {
		defer resp.Body.Close()
		scanner := bufio.NewScanner(resp.Body)
		for scanner.Scan() {
			l := scanner.Text()
			if m, _ := regexp.MatchString("Route results", l); !m {
				continue
			}
			if i := strings.Index(l, "Route results"); i > 0 {
				l = l[i:]
			}
			l = sanitize(l)
			c <- l
		}
		close(c)
	}()
	return c
}

//FetchNodes returns all available nodes through HTTP
func (p *Level3) FetchNodes() map[string]string {
	var nodes = make(map[string]string, 100)
	resp, err := http.Get("http://lookingglass.level3.net/ping/lg_ping_main.php")
	if err != nil {
		println("error: level3 looking glass unreachable (1)")
		return map[string]string{}
	}
	defer resp.Body.Close()
	body, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		println("error: level3 looking glass unreachable (2)" + err.Error())
		return map[string]string{}
	}
	r, _ := regexp.Compile(`(?i)<option value="(?s)([\w|\s|)(._-]+)">(?s)([a-z|\s|)(,._-]+)</option>`)
	b := r.FindAllStringSubmatch(string(body), -1)
	for _, v := range b {
		nodes[v[2]] = v[1]
	}
	return nodes
}
