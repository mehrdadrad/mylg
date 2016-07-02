package dns

import (
	"bufio"
	"fmt"
	"github.com/mehrdadrad/mylg/cli"
	"github.com/mehrdadrad/mylg/data"
	"github.com/miekg/dns"
	"net/http"
	"regexp"
	"strings"
)

type DNS struct {
	host    string
	servers map[string]DNSHost
}

type DNSHost struct {
	IP      string
	Country string
	City    string
}

// Create a new dns object
func NewRequest() *DNS {
	return &DNS{host: ""}
}

// Initialize dns object
func (d *DNS) Init(c *cli.Readline) {
	c.SetPrompt("dns")
	c.Refresh()
	var (
		items     = make(map[string]struct{})
		countries []string
		r, _      = regexp.Compile(`^(\w{2})`)
	)
	sl := fetchDNSHosts()
	d.servers = sl
	for item, _ := range sl {
		i := r.FindStringSubmatch(item)
		if len(i) > 0 {
			items[i[0]] = struct{}{}
		}
	}
	for alpha2 := range items {
		if name, ok := data.Country[alpha2]; ok {
			countries = append(countries, name)
		}
	}
	c.UpdateCompleter("connect", countries)
}

// Look up name server
func (d *DNS) dnsLookup() {
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

// Fetch dns servers from public-dns.info
func fetchDNSHosts() map[string]DNSHost {
	var list = map[string]DNSHost{}
	resp, err := http.Get("http://public-dns.info/nameservers.csv")
	if err != nil {
		println(err.Error())
	}
	defer resp.Body.Close()
	scanner := bufio.NewScanner(resp.Body)
	for scanner.Scan() {
		csv := strings.Split(scanner.Text(), ",")
		if csv[3] != "\"\"" {
			list[csv[2]+" "+csv[3]] = DNSHost{IP: csv[0], Country: csv[2], City: csv[3]}
		}
	}
	return list
}
