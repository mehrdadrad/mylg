// Package ripe provides ASN and IP information
package ripe

import (
	"encoding/json"
	"io/ioutil"
	"net/http"
)

// ASN represents ASN information
type ASN struct {
	Number string
	Data   map[string]interface{}
}

// GetData gets ASN information from RIPE NCC
func (a *ASN) GetData() {
	resp, err := http.Get("https://stat.ripe.net/data/as-overview/data.json?resource=AS" + a.Number)
	if err != nil {
		println(err)
	}
	defer resp.Body.Close()
	body, _ := ioutil.ReadAll(resp.Body)
	json.Unmarshal(body, &a.Data)
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
