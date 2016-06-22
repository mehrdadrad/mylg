package icmp

import (
	"golang.org/x/net/icmp"
	"golang.org/x/net/ipv4"
	"golang.org/x/net/ipv6"
	"log"
	"math/rand"
	"net"
	"sync"
	"time"
)

const (
	ProtocolIPV4ICMP = 1
	ProtocolIPv6ICMP = 58
)

type Ping struct {
	m         icmp.Message
	id        int
	seq       int
	addrs     map[string]*net.IPAddr
	isV4Avail bool
	isV6Avail bool
	network   string
	source    string
	MaxRTT    time.Duration
	mu        sync.RWMutex
}

func NewPing() *Ping {
	return &Ping{
		id:        rand.Intn(0xffff),
		seq:       rand.Intn(0xffff),
		addrs:     make(map[string]*net.IPAddr),
		isV4Avail: false,
		isV6Avail: false,
		network:   "ip",
		source:    "",
		MaxRTT:    time.Second,
	}
}

func isIPv4(ip net.IP) bool {
	return len(ip.To4()) == net.IPv4len
}

func isIPv6(ip net.IP) bool {
	return len(ip) == net.IPv6len
}

func (p *Ping) AddIP(ipAddr string) {
	ip := net.ParseIP(ipAddr)
	p.mu.Lock()
	defer p.mu.Unlock()
	p.addrs[ip.String()] = &net.IPAddr{IP: ip}
}

func (p *Ping) DelIP(ipAddr string) {

}

func (p *Ping) listen(network string) *icmp.PacketConn {
	c, err := icmp.ListenPacket(network, p.source)
	if err != nil {
		log.Fatalf("listen err, %s", err)
	}
	defer c.Close()

	return c
}

func (p *Ping) send(conn *icmp.PacketConn) {
	var (
		wg sync.WaitGroup
	)
	for _, dest := range p.addrs {
		var icmpType icmp.Type
		if isIPv4(dest.IP) {
			icmpType = ipv4.ICMPTypeEcho
		} else if isIPv6(dest.IP) {
			icmpType = ipv6.ICMPTypeEchoRequest
		} else {
			continue
		}

		m := icmp.Message{
			Type: icmpType, Code: 0,
			Body: &icmp.Echo{
				ID:   p.id,
				Seq:  1,
				Data: []byte("myping"),
			},
		}
		bytes, err := m.Marshal(nil)
		if err != nil {

		}
		wg.Add(1)
		go func(conn *icmp.PacketConn, dest net.Addr, b []byte) {

		}(conn, dest, bytes)

	}
}

func (p *Ping) start() {
	//var conn, conn6 *icmp.PacketConn

}

func (p *Ping) Start() {

}
