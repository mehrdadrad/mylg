// Package ns provides name server methods for selected name server(s)
package ns

import (
	"bufio"
	"bytes"
	"fmt"
	"io/ioutil"
	"net/http"
	"os"
	"sort"
	"strings"
	"time"

	"github.com/mehrdadrad/mylg/cli"
	"github.com/mehrdadrad/mylg/data"
	"github.com/miekg/dns"
)

// A Host represents a name server host
type Host struct {
	ip      string
	alpha2  string
	country string
	city    string
}

// A Request represents a name server request
type Request struct {
	country string
	host    string
	hosts   []Host
}

// NewRequest creates a new dns request object
func NewRequest() *Request {
	return &Request{host: ""}
}

// Init configure dns command and fetch name servers
func (d *Request) Init() {
	if !d.cache("validate") {
		d.hosts = fetchNSHosts()
		d.cache("write")
	} else {
		d.cache("read")
	}
}

// SetCountryList init the connect contry items
func (d *Request) SetCountryList(c *cli.Readline) {
	var countries []string
	for _, host := range d.hosts {
		countries = append(countries, host.country)
	}
	countries = uniqStrSlice(countries)
	sort.Strings(countries)
	c.UpdateCompleter("connect", countries)
}

// SetNodeList init the node city items
func (d *Request) SetNodeList(c *cli.Readline) {
	var node []string
	for _, host := range d.hosts {
		if host.country == d.country {
			node = append(node, host.city)
		}
	}
	sort.Strings(node)
	c.UpdateCompleter("node", node)
}

// ValidateCountry set requested country
func (d *Request) ValidateCountry(args string) bool {
	d.country = args
	return true
}

// Look up name server
func (d *Request) dnsLookup() {
	//var list []DNSHost

	c := new(dns.Client)
	m := new(dns.Msg)

	m.SetQuestion(dns.Fqdn("yahoo.com"), dns.TypeA)
	m.RecursionDesired = true
	r, _, err := c.Exchange(m, "8.8.8.8:53")
	if err != nil {
		println(err.Error())
	}
	for _, a := range r.Answer {
		fmt.Printf("%#v\n", a)
	}
}

// cache provides caching for name servers
func (d *Request) cache(r string) bool {
	switch r {
	case "read":
		b, err := ioutil.ReadFile("/tmp/mylg.ns")
		if err != nil {
			panic(err.Error())
		}
		d.hosts = d.hosts[:0]
		r := bytes.NewBuffer(b)
		s := bufio.NewScanner(r)
		for s.Scan() {
			csv := strings.Split(s.Text(), ";")
			if len(csv) != 4 {
				continue
			}
			d.hosts = append(d.hosts, Host{alpha2: csv[0], country: csv[1], city: csv[2], ip: csv[3]})
		}
	case "write":
		var data []string
		for _, h := range d.hosts {
			data = append(data, fmt.Sprintf("%s;%s;%s;%s", h.alpha2, h.country, h.city, h.ip))
		}
		err := ioutil.WriteFile("/tmp/mylg.ns", []byte(strings.Join(data, "\n")), 0644)
		if err != nil {
			panic(err.Error())
		}
		println("write done!")
	case "validate":
		f, err := os.Stat("/tmp/mylg.ns")
		if err != nil {
			return false
		}
		d := time.Since(f.ModTime())
		if d.Hours() > 48 {
			return false
		}
	}
	return true
}

// Fetch name servers from public-dns.info
func fetchNSHosts() []Host {
	var (
		hosts   []Host
		counter = make(map[string]int)
	)
	resp, err := http.Get("http://public-dns.info/nameservers.csv")
	if err != nil {
		println(err.Error())
	}
	defer resp.Body.Close()
	scanner := bufio.NewScanner(resp.Body)
	for scanner.Scan() {
		csv := strings.Split(scanner.Text(), ",")
		if csv[3] != "\"\"" {
			if name, ok := data.Country[csv[2]]; ok && counter[csv[2]] < 5 {
				hosts = append(hosts, Host{ip: csv[0], alpha2: csv[2], country: name, city: csv[3]})
				counter[csv[2]]++
			}
		}
	}
	return hosts
}

// uniqStrSlice return unique slice
func uniqStrSlice(src []string) []string {
	var rst []string
	tmp := make(map[string]struct{})
	for _, s := range src {
		tmp[s] = struct{}{}
	}
	for s := range tmp {
		rst = append(rst, s)
	}
	return rst
}
