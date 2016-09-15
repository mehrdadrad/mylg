package icmp_test

import (
	"github.com/mehrdadrad/mylg/cli"
	"github.com/mehrdadrad/mylg/icmp"
	"net"
	"testing"
)

func TestSetIP(t *testing.T) {
	cfg, _ := cli.ReadDefaultConfig()
	_, err := icmp.NewPing("8.8.8.8", cfg)
	if err != nil {
		t.Error("NewPing failed with error:", err)
	}
}

func TestIsIPvx(t *testing.T) {
	ip := net.ParseIP("8.8.8.8")
	if !icmp.IsIPv4(ip) {
		t.Error("IsIPv4 is false but expected true")
	}
}
