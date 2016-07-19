package ns_test

import (
	"testing"

	"github.com/mehrdadrad/mylg/ns"
)

var r *ns.Request

func init() {
	r = ns.NewRequest()
	r.Hosts = append(r.Hosts, ns.Host{"127.0.0.1", "us", "united states", "los angeles"})
}

func TestNodeList(t *testing.T) {
	r.Country = "united states"
	nodes := r.NodeList()
	if len(nodes) != 1 || nodes[0] != "los angeles" {
		t.Error("NodeList didn't return expected city")
	}
}

func TestCountryList(t *testing.T) {
	r.Country = "united states"
	countries := r.CountryList()
	if len(countries) != 1 || countries[0] != "united states" {
		t.Error("CountryList didn't return expected country")
	}
}

func TestChkCountry(t *testing.T) {
	if !r.ChkCountry("united states") {
		t.Error("ChkCountry didn't return expected value")
	}
}

func TestChkNode(t *testing.T) {
	if !r.ChkNode("los angeles") {
		t.Error("ChkNode didn't return expected value")
	}
}
