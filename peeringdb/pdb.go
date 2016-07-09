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

// A Peer represents peeringdb record
type Peer struct {
	Name    string `json:"name"`
	ASN     int    `json:"asn"`
	Status  string `json:"status"`
	Speed   int    `json:"speed"`
	IPAddr4 string `json:"ipaddr4"`
	IPAddr6 string `json:"ipaddr6"`
}

// Peers represents peeringdb records
type Peers struct {
	Data []Peer `json:"data"`
}

// getData fetchs netixlan data from peeringdb
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

// printTable prints peeringdb data as table
func printTable(data [][]string) {
	table := tablewriter.NewWriter(os.Stdout)
	table.SetHeader([]string{"Name", "Status", "Speed", "IPv4 Addr", "IPv6 Addr"})
	for _, v := range data {
		table.Append(v)
	}
	table.Render()
}

// Search find a key through the records
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

// isASN checks if the key is number
func isASN(key string) bool {
	m, err := regexp.MatchString(`^(?i)\d{2,5}`, key)
	if err != nil {
		return false
	}
	return m
}
