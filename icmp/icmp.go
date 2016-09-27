// Copyright 2016 Mehrdad Arshad Rad <arshad.rad@gmail.com>. All rights reserved.
// Use of this source code is governed by a MIT license that can
// be found in the LICENSE file.

package icmp

import (
	"net"
	"os"
	"syscall"
	"time"

	"golang.org/x/net/icmp"
	"golang.org/x/net/ipv4"
	"golang.org/x/net/ipv6"
)

const (
	// DefaultTXTimeout is socket send timeout
	DefaultTXTimeout int64 = 2000
	// ProtocolIPv4ICMP is IANA ICMP IPv4
	ProtocolIPv4ICMP = 1
	// ProtocolIPv6ICMP is IANA ICMP IPv6
	ProtocolIPv6ICMP = 58

	// IPv4ICMPTypeEchoReply is ICMPv4 Echo Reply
	IPv4ICMPTypeEchoReply = 0
	// IPv4ICMPTypeDestinationUnreachable is ICMPv4 Destination Unreachable
	IPv4ICMPTypeDestinationUnreachable = 3
	// IPv4ICMPTypeTimeExceeded is ICMPv4 Time Exceeded
	IPv4ICMPTypeTimeExceeded = 11

	// IPv6ICMPTypeEchoReply is ICMPv6 Echo Reply
	IPv6ICMPTypeEchoReply = 129
	// IPv6ICMPTypeDestinationUnreachable is ICMPv6 Destination Unreachable
	IPv6ICMPTypeDestinationUnreachable = 1
	//IPv6ICMPTypeTimeExceeded is ICMPv6 Time Exceeded
	IPv6ICMPTypeTimeExceeded = 3
)

// Trace represents trace properties
type Trace struct {
	src      string
	host     string
	ip       net.IP
	ips      []net.IP
	ttl      int
	fd       int
	family   int
	proto    int
	wait     string
	icmp     bool
	resolve  bool
	ripe     bool
	realTime bool
	pSize    int
	maxTTL   int

	uiTheme string
}

// Ping represents ping request
type Ping struct {
	m         icmp.Message
	id        int
	seq       int
	pSize     int
	count     int
	addr      *net.IPAddr
	addrs     []net.IP
	target    string
	isV4Avail bool
	isV6Avail bool
	isCIDR    bool
	forceV4   bool
	forceV6   bool
	network   string
	source    string
	timeout   time.Duration
	interval  time.Duration
	MaxRTT    time.Duration
}

// HopResp represents hop's response
type HopResp struct {
	num     int
	hop     string
	ip      string
	elapsed float64
	last    bool
	err     error
	whois   Whois
}

// ICMPResp represents ICMP response msg
type ICMPResp struct {
	typ  int
	code int
	id   int
	seq  int
	src  net.IP
	udp  struct{ dstPort int }
	ip   struct{ dst net.IP }
}

// Whois represents prefix info from RIPE
type Whois struct {
	holder string
	asn    float64
}

// Stats represents statistic's fields
type Stats struct {
	count int64   // sent packet
	min   float64 // minimum/best rtt
	avg   float64 // average rtt
	max   float64 // maximum/worst rtt
	pkl   int64   // packet loss
}

func bytesToIPv6(b []byte) net.IP {
	ip := make(net.IP, net.IPv6len)
	copy(ip, b)
	return ip
}

func icmpV4RespParser(b []byte) ICMPResp {
	var resp ICMPResp

	resp.typ = int(b[20])
	resp.code = int(b[21])
	resp.src = net.IPv4(b[12], b[13], b[14], b[15])

	switch resp.typ {
	case IPv4ICMPTypeEchoReply:
		resp.id = int(b[24])<<8 | int(b[25])
		resp.seq = int(b[26])<<8 | int(b[27])
	case IPv4ICMPTypeDestinationUnreachable:
		resp.ip.dst = net.IPv4(b[44], b[45], b[46], b[47])
	case IPv4ICMPTypeTimeExceeded:
		resp.id = int(b[52])<<8 | int(b[53])
		resp.seq = int(b[54])<<8 | int(b[55])
		resp.ip.dst = net.IPv4(b[44], b[45], b[46], b[47])
	}

	return resp
}

func icmpV6RespParser(b []byte) ICMPResp {
	var resp ICMPResp

	resp.typ = int(b[0])
	resp.code = int(b[1])

	//getting time exceeded w/ less than 32 bytes
	if len(b) < 44 {
		return resp
	}

	switch resp.typ {
	case IPv6ICMPTypeEchoReply:
		resp.id = int(b[44])<<8 | int(b[25])
		resp.seq = int(b[46])<<8 | int(b[27])
	case IPv6ICMPTypeDestinationUnreachable:
		resp.ip.dst = bytesToIPv6(b[32:48])
	case IPv6ICMPTypeTimeExceeded:
		resp.id = int(b[32])<<8 | int(b[33])
		resp.seq = int(b[34])<<8 | int(b[35])
		resp.ip.dst = bytesToIPv6(b[32:48])
	}

	return resp
}

func icmpV6Message(id, seq int) ([]byte, error) {
	m, err := (&icmp.Message{
		Type: ipv6.ICMPTypeEchoRequest, Code: 0,
		Body: &icmp.Echo{
			ID: id, Seq: seq,
			Data: []byte("myLG - [mylg.io]"),
		},
	}).Marshal(nil)

	if err != nil {
		return m, os.NewSyscallError("icmpmsg", err)
	}
	return m, nil
}

func icmpV4Message(id, seq int) ([]byte, error) {
	m, err := (&icmp.Message{
		Type: ipv4.ICMPTypeEcho, Code: 0,
		Body: &icmp.Echo{
			ID: id, Seq: seq,
			Data: []byte("myLG - [mylg.io]"),
		},
	}).Marshal(nil)

	if err != nil {
		return m, os.NewSyscallError("icmpmsg", err)
	}
	return m, nil
}

func setIPv4TOS(fd int, v int) error {
	err := syscall.SetsockoptInt(fd, syscall.IPPROTO_IP, syscall.IP_TOS, v)
	if err != nil {
		return os.NewSyscallError("setsockopt", err)
	}
	return nil
}

func setIPv4TTL(fd int, v int) error {
	err := syscall.SetsockoptInt(fd, syscall.IPPROTO_IP, syscall.IP_TTL, v)
	if err != nil {
		return os.NewSyscallError("setsockopt", err)
	}
	return nil
}

func setIPv6HopLimit(fd int, v int) error {
	err := syscall.SetsockoptInt(fd, syscall.IPPROTO_IPV6, syscall.IPV6_UNICAST_HOPS, v)
	if err != nil {
		return os.NewSyscallError("setsockopt", err)
	}
	return nil
}
