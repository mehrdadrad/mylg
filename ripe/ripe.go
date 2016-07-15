// Package ripe provides ASN and IP information
package ripe

import (
	"encoding/json"
	"io/ioutil"
	"net/http"
	"regexp"
)

const (
	RIPEAPI       = "https://stat.ripe.net"
	RIPEPrefixURL = "/data/prefix-overview/data.json?max_related=50&resource="
	RIPEASNURL    = "/data/as-overview/data.json?resource=AS"
)

// ASN represents ASN information
type ASN struct {
	Number string
	Data   map[string]interface{}
}

// Prefix represents prefix information
type Prefix struct {
	Resource string
	Data     map[string]interface{}
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
	} else {
		println("error")
	}
}

// Set ASN
func (a *ASN) Set(r string) {
	a.Number = r
}

// GetData gets ASN information from RIPE NCC
func (a *ASN) GetData() bool {
	if len(a.Number) < 2 {
		println("error: AS number invalid")
		return false
	}
	resp, err := http.Get(RIPEAPI + RIPEASNURL + a.Number)
	if err != nil {
		println(err)
		return false
	}
	if resp.StatusCode != 200 {
		println("error: check your AS number")
		return false
	}
	defer resp.Body.Close()
	body, _ := ioutil.ReadAll(resp.Body)
	json.Unmarshal(body, &a.Data)
	return true
}

// PrettyPrint print ASN information (holder)
func (a *ASN) PrettyPrint() {
	data, ok := a.Data["data"].(map[string]interface{})
	if ok {
		println(string(data["holder"].(string)))
	} else {
		println("error")
	}
}

// IsASN checks if the key is a number
func IsASN(key string) bool {
	m, err := regexp.MatchString(`^\d+$`, key)
	if err != nil {
		return false
	}
	return m
}
