package peeringdb

import (
	"encoding/json"
	"fmt"
	"github.com/olekukonko/tablewriter"
	"io/ioutil"
	"net/http"
	"os"
	"regexp"
	"strconv"
	"time"

	"github.com/mehrdadrad/mylg/cli"
)

const (
	// APINetIXLAN holds net ix lan API
	APINetIXLAN = "https://www.peeringdb.com/api/netixlan"
	// APINet holds net API
	APINet = "https://www.peeringdb.com/api/net"
)

// A Peer represents peeringdb record
type Peer struct {
	Name    string `json:"name"`
	ASN     int    `json:"asn"`
	Status  string `json:"status"`
	Speed   int    `json:"speed"`
	IPAddr4 string `json:"ipaddr4"`
	IPAddr6 string `json:"ipaddr6"`
}

// A Net represents peeringdb net record
type Net struct {
	Name     string `json:"name"`
	ASN      int    `json:"asn"`
	WWW      string `json:"website"`
	Traffic  string `json:"info_traffic"`
	InfoType string `json:"info_type"`
	Note     string `json:"notes"`
}

// Peers represents peeringdb records
type Peers struct {
	Data []Peer `json:"data"`
}

// Nets represents network information data
type Nets struct {
	Data []Net `json:"data"`
}

// GetNetIXLAN fetchs netixlan data from peeringdb
func GetNetIXLAN() (interface{}, error) {
	var peers Peers
	resp, err := http.Get(APINetIXLAN)
	if err != nil {
		return peers, fmt.Errorf("peeringdb.com is unreachable (1)")
	}
	if resp.StatusCode != 200 {
		return peers, fmt.Errorf("peeringdb.com returns none 200 HTTP code")
	}
	defer resp.Body.Close()
	body, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		return peers, fmt.Errorf("peeringdb.com is unreachable (2)  %s", err.Error())
	}
	json.Unmarshal(body, &peers)
	return peers, nil
}

// GetNet fetchs net information from peeringdb
func GetNet() (interface{}, error) {
	var (
		nets Nets
		res  = make(map[string]Net)
	)
	resp, err := http.Get(APINet)
	if err != nil {
		return nil, fmt.Errorf("peeringdb.com is unreachable (1)")
	}
	if resp.StatusCode != 200 {
		return nil, fmt.Errorf("peeringdb.com returns none 200 HTTP code")
	}
	defer resp.Body.Close()
	body, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		return nil, fmt.Errorf("peeringdb.com is unreachable (2)  %s", err.Error())
	}
	json.Unmarshal(body, &nets)
	for _, n := range nets.Data {
		res[strconv.Itoa(n.ASN)] = n
	}
	return res, nil
}

// cache manage caching for peeringdb data
func cache(r, typ string, data interface{}) (interface{}, bool) {
	switch r {
	case "write":
		b, err := json.Marshal(data)
		if err != nil {
			return nil, false
		}
		err = ioutil.WriteFile("/tmp/mylg.pdb."+typ, b, 0644)
		if err != nil {
			return nil, false
		}
	case "read":
		b, err := ioutil.ReadFile("/tmp/mylg.pdb." + typ)
		if err != nil {
			return nil, false
		}
		if typ == "ix" {
			var res Peers
			err := json.Unmarshal(b, &res)
			if err != nil {
				return nil, false
			}
			return res, true
		}
		// net type
		var res map[string]Net
		err = json.Unmarshal(b, &res)
		if err != nil {
			return nil, false
		}
		return res, true
	case "validate":
		f, err := os.Stat("/tmp/mylg.pdb." + typ)
		if err != nil {
			return nil, false
		}
		d := time.Since(f.ModTime())
		if d.Hours() > 96 {
			return nil, false
		}
	}

	return nil, true
}

// printTable prints peeringdb data as table
func printTable(net Net, ixLan [][]string) {
	println("Data provided by www.peeringdb.com")
	table := tablewriter.NewWriter(os.Stdout)
	table.SetHeader([]string{"Name", "Traffic", "Type", "Web site", "Note"})
	table.Append([]string{net.Name, net.Traffic, net.InfoType, net.WWW, net.Note})
	table.Render()
	table = tablewriter.NewWriter(os.Stdout)
	table.SetHeader([]string{"Name", "Status", "Speed", "IPv4 Addr", "IPv6 Addr"})
	for _, v := range ixLan {
		table.Append(v)
	}
	table.Render()
}

// Search find a key through the records
func Search(key string) {
	var (
		result [][]string
		ASN    string
		peers  interface{}
		nets   interface{}
		err    error
	)

	key, flag := cli.Flag(key)
	// help
	if _, ok := flag["help"]; ok {
		help()
		return
	}

	if _, ok := cache("validate", "ix", nil); ok {
		peers, _ = cache("read", "ix", nil)
	} else {
		peers, err = GetNetIXLAN()
		cache("write", "ix", peers)
	}
	if _, ok := cache("validate", "net", nil); ok {
		nets, _ = cache("read", "net", nil)
	} else {
		nets, err = GetNet()
		cache("write", "net", nets)
	}
	if err != nil {
		println(err.Error())
		return
	}

	switch {
	case IsASN(key):
		ASN = key
		for _, p := range peers.(Peers).Data {
			if strconv.Itoa(p.ASN) == key {
				result = append(result, []string{p.Name, p.Status, strconv.Itoa(p.Speed), p.IPAddr4, p.IPAddr6})
			}
		}
	}
	if len(result) > 0 {
		n := nets.(map[string]Net)
		printTable(n[ASN], result)
	} else {
		println("no information @ peeringdb")
	}
}

// IsASN checks if the key is number
func IsASN(key string) bool {
	m, err := regexp.MatchString(`^(?i)\d{2,5}`, key)
	if err != nil {
		return false
	}
	return m
}

// help represents guide to user
func help() {
	println(`
    usage:
          peering AS.Number
    example:
          peering 577
	`)
}
