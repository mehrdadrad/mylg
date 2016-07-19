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

	"github.com/mehrdadrad/mylg/data"
	"github.com/miekg/dns"
)

const (
	publicDNSHost      = "http://public-dns.info"
	publicDNSNodesPath = "/nameservers.csv"
)

// A Host represents a name server host
type Host struct {
	IP      string
	Alpha2  string
	Country string
	City    string
}

// A Request represents a name server request
type Request struct {
	Country string
	City    string
	Host    string
	Hosts   []Host
}

// NewRequest creates a new dns request object
func NewRequest() *Request {
	return &Request{Host: ""}
}

// Init configure dns command and fetch name servers
func (d *Request) Init() {
	if !d.cache("validate") {
		d.Hosts = fetchNSHosts()
		d.cache("write")
	} else {
		d.cache("read")
	}
}

// CountryList init the connect contry items
func (d *Request) CountryList() []string {
	var countries []string
	for _, host := range d.Hosts {
		countries = append(countries, host.Country)
	}
	countries = uniqStrSlice(countries)
	sort.Strings(countries)
	return countries
}

// NodeList gets the node city items
func (d *Request) NodeList() []string {
	var node []string
	for _, host := range d.Hosts {
		if host.Country == d.Country {
			node = append(node, host.City)
		}
	}
	sort.Strings(node)
	return node
}

// ChkCountry validates and set requested country
func (d *Request) ChkCountry(country string) bool {
	country = strings.ToLower(country)
	for _, h := range d.Hosts {
		if country == h.Country {
			d.Country = country
			return true
		}
	}

	return false
}

// ChkNode set requested country
func (d *Request) ChkNode(city string) bool {
	city = strings.ToLower(city)
	for _, h := range d.Hosts {
		if d.Country == h.Country && city == h.City {
			d.Host = h.IP
			d.City = h.City
			return true
		}
	}
	return false
}

// Local set host to nothing means local
func (d *Request) Local() {
	d.Host = ""
	d.Country = ""
}

// Dig look up name server
func (d *Request) Dig(args string) {
	c := new(dns.Client)
	m := new(dns.Msg)

	m.SetQuestion(dns.Fqdn(args), dns.TypeANY)
	m.RecursionDesired = true

	if d.Host == "" {
		config, _ := dns.ClientConfigFromFile("/etc/resolv.conf")
		d.Host = config.Servers[0]
		d.City = "your local dns server"
	}

	println("Trying to query server:", d.Host, d.Country, d.City)

	t := time.Now()
	r, _, err := c.Exchange(m, d.Host+":53")
	elapsed := time.Now().Sub(t)
	if err != nil {
		println(err.Error())
		return
	}
	fmt.Printf("Query time: %.4f ms\n", elapsed.Seconds())
	for _, a := range r.Answer {
		fmt.Println(a)
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
		d.Hosts = d.Hosts[:0]
		r := bytes.NewBuffer(b)
		s := bufio.NewScanner(r)
		for s.Scan() {
			csv := strings.Split(s.Text(), ";")
			if len(csv) != 4 {
				continue
			}
			d.Hosts = append(d.Hosts, Host{Alpha2: csv[0], Country: csv[1], City: csv[2], IP: csv[3]})
		}
	case "write":
		var data []string
		for _, h := range d.Hosts {
			data = append(data, fmt.Sprintf("%s;%s;%s;%s", h.Alpha2, h.Country, h.City, h.IP))
		}
		err := ioutil.WriteFile("/tmp/mylg.ns", []byte(strings.Join(data, "\n")), 0644)
		if err != nil {
			panic(err.Error())
		}
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
		city    string
		counter = make(map[string]int)
		chkDup  = make(map[string]int)
	)
	resp, err := http.Get(publicDNSHost + publicDNSNodesPath)
	if err != nil {
		println(err.Error())
		return []Host{}
	}
	if resp.StatusCode != 200 {
		println("error: public dns is not available")
		return []Host{}
	}
	defer resp.Body.Close()
	scanner := bufio.NewScanner(resp.Body)
	for scanner.Scan() {
		csv := strings.Split(scanner.Text(), ",")
		if csv[3] != "\"\"" {
			if name, ok := data.Country[csv[2]]; ok && counter[csv[2]] < 5 {
				name = strings.ToLower(name)
				city = strings.ToLower(csv[3])
				chkDup[name+city] += 1
				if chkDup[name+city] > 1 {
					city = fmt.Sprintf("%s0%d", city, chkDup[name+city]-1)
				}

				hosts = append(hosts, Host{IP: csv[0], Alpha2: csv[2], Country: name, City: city})
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
