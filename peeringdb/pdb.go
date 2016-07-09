package peeringdb

import (
	"encoding/json"
	_ "fmt"
	"github.com/olekukonko/tablewriter"
	"io/ioutil"
	"net/http"
	"os"
	"regexp"
	"strconv"
)

type Peer struct {
	Name    string `json:"name"`
	ASN     int    `json:"asn"`
	Status  string `json:"status"`
	Speed   int    `json:"speed"`
	IPAddr4 string `json:"ipaddr4"`
	IPAddr6 string `json:"ipaddr6"`
}
type Peers struct {
	Data []Peer `json:"data"`
}

func getData() Peers {
	var peers Peers
	resp, err := http.Get("https://www.peeringdb.com/api/netixlan")
	if err != nil {
		println("error: peeringdb.com is unreachable (1) ")
	}
	defer resp.Body.Close()
	body, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		println("error: peeringdb.com is unreachable (2)" + err.Error())
	}
	json.Unmarshal(body, &peers)
	return peers
}

func printTable(data [][]string) {
	table := tablewriter.NewWriter(os.Stdout)
	table.SetHeader([]string{"Name", "Status", "Speed", "IPv4 Addr", "IPv6 Addr"})
	for _, v := range data {
		table.Append(v)
	}
	table.Render()
}

func Search(key string) {
	var result [][]string
	peers := getData()
	switch {
	case isASN(key):
		for _, p := range peers.Data {
			if strconv.Itoa(p.ASN) == key {
				result = append(result, []string{p.Name, p.Status, strconv.Itoa(p.Speed), p.IPAddr4, p.IPAddr6})
			}
		}
	}

	if len(result) > 0 {
		printTable(result)
	} else {
		println("there is not any information @ peeringdb")
	}
}

func isASN(key string) bool {
	m, err := regexp.MatchString(`^(?i)\d{2,5}`, key)
	if err != nil {
		return false
	}
	return m
}

/*
func cache(r string) string, bool {
	var fn = "/tmp/mylg.pdb"
	switch r {
	case "read":
		b, err := ioutil.ReadFile(fn)
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
		err := ioutil.WriteFile(fn, []byte(strings.Join(data, "\n")), 0644)
		if err != nil {
			panic(err.Error())
		}
	case "validate":
		f, err := os.Stat(fn)
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
*/
