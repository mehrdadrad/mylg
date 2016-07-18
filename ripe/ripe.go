// Package ripe provides ASN and IP information
package ripe

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
	"net/http"
	"os"
	"regexp"
	"sort"
	"strings"
	"sync"

	"github.com/mehrdadrad/mylg/data"
	"github.com/olekukonko/tablewriter"
)

const (
	// Ripe API URL
	RIPEAPI = "https://stat.ripe.net"
	// Ripe prefix path
	RIPEPrefixURL = "/data/prefix-overview/data.json?max_related=50&resource="
	// Ripe ASN path
	RIPEASNURL = "/data/as-overview/data.json?resource=AS"
	// Ripe Geo path
	RIPEGeoURL = "/data/geoloc/data.json?resource=AS"
)

// ASN represents ASN information
type ASN struct {
	Number  string
	Data    map[string]interface{}
	GeoData map[string]interface{}
}

// Prefix represents prefix information
type Prefix struct {
	Resource string
	Data     map[string]interface{}
}

// kv represents key/value(float64) in sort func
type kv struct {
	key   string
	value float64
}

// location represents location information
type location struct {
	City    string `json:"city"`
	Country string `json:"country"`
}

// Set sets the resource value
func (p *Prefix) Set(r string) {
	p.Resource = r
}

// GetData gets prefix information from RIPE NCC
func (p *Prefix) GetData() bool {
	if len(p.Resource) < 6 {
		println("error: prefix invalid")
		return false
	}
	resp, err := http.Get(RIPEAPI + RIPEPrefixURL + p.Resource)
	if err != nil {
		println(err.Error())
		return false
	}
	if resp.StatusCode != 200 {
		println("error: check your prefix")
		return false
	}
	defer resp.Body.Close()
	body, _ := ioutil.ReadAll(resp.Body)
	json.Unmarshal(body, &p.Data)
	return true
}

// PrettyPrint print ASN information (holder)
func (p *Prefix) PrettyPrint() {
	data, ok := p.Data["data"].(map[string]interface{})
	if ok {
		println("prefix:", data["resource"].(string))
		asns := data["asns"].([]interface{})
		for _, h := range asns {
			println("holder:", h.(map[string]interface{})["holder"].(string))
		}
	}
}

// Set ASN
func (a *ASN) Set(r string) {
	a.Number = r
}

// GetData gets ASN information from RIPE NCC
func (a *ASN) GetData() bool {
	var (
		wg        sync.WaitGroup
		rOV, rGeo bool
	)

	wg.Add(1)
	go func() {
		defer wg.Done()
		rOV = a.GetOVData()
	}()

	wg.Add(1)
	go func() {
		defer wg.Done()
		rGeo = a.GetGeoData()
	}()
	wg.Wait()
	return rOV || rGeo
}

// GetOVData gets ASN overview information from RIPE NCC
func (a *ASN) GetOVData() bool {
	if len(a.Number) < 2 {
		println("error: AS number invalid")
		return false
	}
	resp, err := http.Get(RIPEAPI + RIPEASNURL + a.Number)
	if err != nil {
		println(err.Error())
		return false
	}
	if resp.StatusCode != 200 {
		return false
	}
	defer resp.Body.Close()
	body, _ := ioutil.ReadAll(resp.Body)
	json.Unmarshal(body, &a.Data)
	return true
}

// GetGeoData gets Geo information from RIPE NCC
func (a *ASN) GetGeoData() bool {
	if len(a.Number) < 2 {
		println("error: AS number invalid")
		return false
	}
	resp, err := http.Get(RIPEAPI + RIPEGeoURL + a.Number)
	if err != nil {
		println(err.Error())
		return false
	}
	if resp.StatusCode != 200 {
		println("error: check your AS number")
		return false
	}
	defer resp.Body.Close()
	body, _ := ioutil.ReadAll(resp.Body)
	json.Unmarshal(body, &a.GeoData)
	return true
}

// PrettyPrint print ASN information (holder)
func (a *ASN) PrettyPrint() {
	var cols = make(map[string]float64)
	overviewData, ok := a.Data["data"].(map[string]interface{})
	if ok {
		println(string(overviewData["holder"].(string)))
	}
	geoLocData, ok := a.GeoData["data"].(map[string]interface{})
	if !ok {
		return
	}
	locs := geoLocData["locations"].([]interface{})
	table := tablewriter.NewWriter(os.Stdout)
	table.SetHeader([]string{"Location", "Covered %"})
	for _, loc := range locs {
		geoInfo := loc.(map[string]interface{})
		cols[geoInfo["country"].(string)] = geoInfo["covered_percentage"].(float64)
	}
	for _, v := range sortMapFloat(cols) {
		name := v.key
		percent := v.value
		uc := strings.Split(name, "-")
		if country, ok := data.Country[uc[0]]; ok {
			name = country
		}
		if len(uc) == 2 {
			name = fmt.Sprintf("%s - %s", name, uc[1])
		}
		table.Append([]string{name, fmt.Sprintf("%.2f", percent)})
	}
	table.Render()
}

// IsASN checks if the key is a number
func IsASN(key string) bool {
	m, err := regexp.MatchString(`^\d+$`, key)
	if err != nil {
		return false
	}
	return m
}

// sortMapFloat sorts map[string]float64 w/ value
func sortMapFloat(m map[string]float64) []kv {
	n := map[float64][]string{}
	var (
		a []float64
		r []kv
	)
	for k, v := range m {
		n[v] = append(n[v], k)
	}
	for k := range n {
		a = append(a, k)
	}
	sort.Sort(sort.Reverse(sort.Float64Slice(a)))
	for _, k := range a {
		for _, s := range n[k] {
			r = append(r, kv{s, k})
		}
	}
	return r
}
