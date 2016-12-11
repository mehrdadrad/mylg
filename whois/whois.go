// Package whois tries to get information about
// IP address / Prefix / Domain (todo)
package whois

import (
	"github.com/mehrdadrad/mylg/ripe"
)

// whois represents whois providers
type whois interface {
	Set(r string)
	GetData() bool
	PrettyPrint()
}

var (
	w = map[string]whois{"asn": new(ripe.ASN), "prefix": new(ripe.Prefix)}
)

// Lookup tries to get whois information
// ASN and prefix/ip information
func Lookup(args string) {
	if ripe.IsASN(args) {
		w["asn"].Set(args)
		w["asn"].GetData()
		w["asn"].PrettyPrint()
	} else if ripe.IsIP(args) || ripe.IsPrefix(args) {
		w["prefix"].Set(args)
		w["prefix"].GetData()
		w["prefix"].PrettyPrint()
	} else {
		help()
	}
}

// help represents whois help
func help() {
	println(`
    usage:
          whois ASN/CIDR/IPAddress

    Example:
          whois 8.8.8.8
          whois 8.0.0.0/8
          whois 577
	`)
}
