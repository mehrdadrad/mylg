package main

import (
	//"github.com/mehrdadrad/myping/cmd"
	"github.com/mehrdadrad/myping/icmp"
	"net"
)

func resIPAddr(t string, name string) (*net.IPAddr, error) {
	ip, err := net.ResolveIPAddr(t, name)
	return ip, err
}

func main() {
	p := icmp.NewPing()
	p.AddIP(8.8.8.8)
}
