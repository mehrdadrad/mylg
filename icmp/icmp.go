package icmp

import (
	"fmt"
	"golang.org/x/net/icmp"
	"golang.org/x/net/ipv4"
	"golang.org/x/net/ipv6"
	"log"
	"math/rand"
	"net"
	"sync"
	"syscall"
	"time"
)

const (
	ProtocolIPv4ICMP = 1  // IANA ICMP for IPv4
	ProtocolIPv6ICMP = 58 // IANA ICMP for IPv6
)

type packet struct {
	bytes []byte
	addr  net.Addr
}
type Ping struct {
	m         icmp.Message
	id        int
	seq       int
	pSize     int
	addr      *net.IPAddr
	isV4Avail bool
	isV6Avail bool
	network   string
	source    string
	MaxRTT    time.Duration
	mu        sync.RWMutex
}

func resolveHost(t string, name string) (*net.IPAddr, error) {
	ip, err := net.ResolveIPAddr(t, name)
	return ip, err
}
func NewPing() *Ping {
	return &Ping{
		id:        rand.Intn(0xffff),
		seq:       rand.Intn(0xffff),
		pSize:     64,
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

func (p *Ping) ParseHeader(m *packet) (*ipv4.Header, error) {
	var proto int = ProtocolIPv4ICMP
	if p.isV6Avail {
		proto = ProtocolIPv6ICMP
	}

	rm, _ := icmp.ParseMessage(proto, m.bytes)
	bytes, _ := rm.Body.Marshal(proto)
	h, err := icmp.ParseIPv4Header(bytes)

	return h, err
}

func (p *Ping) IP(ipAddr string) {
	ip := net.ParseIP(ipAddr)
	p.mu.Lock()
	defer p.mu.Unlock()
	p.addr = &net.IPAddr{IP: ip}
	if isIPv4(ip) {
		p.isV4Avail = true
	} else {
		p.isV6Avail = true
	}
}

func (p *Ping) DelIP(ipAddr string) {

}

func (p *Ping) Network() {

}

func (p *Ping) SetTTL() {

}
func (p *Ping) PacketSize(s int) {
	p.pSize = s
}

func (p *Ping) listen(network string) *icmp.PacketConn {
	c, err := icmp.ListenPacket(network, p.source)
	if err != nil {
		log.Fatalf("listen err, %s", err)
	}
	return c
}

func (p *Ping) recv(conn *icmp.PacketConn, rcvdChan chan<- *packet) {
	bytes := make([]byte, 512)
	conn.SetReadDeadline(time.Now().Add(time.Millisecond * 100))
	n, dest, err := conn.ReadFrom(bytes)
	if err != nil {

	}
	bytes = bytes[:n]
	select {
	case rcvdChan <- &packet{bytes: bytes, addr: dest}:
	}
}

func (p *Ping) payload() []byte {
	timeBytes := make([]byte, 8)
	ts := time.Now().UnixNano()
	for i := uint8(0); i < 8; i++ {
		timeBytes[i] = byte((ts >> (i * 8)) & 0xff)
	}
	payload := make([]byte, p.pSize-16)
	payload = append(payload, timeBytes...)
	return payload
}
func (p *Ping) send(conn *icmp.PacketConn) {
	var (
		wg sync.WaitGroup
	)
	var icmpType icmp.Type
	if isIPv4(p.addr.IP) {
		icmpType = ipv4.ICMPTypeEcho
	} else if isIPv6(p.addr.IP) {
		icmpType = ipv6.ICMPTypeEchoRequest
	}

	bytes, err := (&icmp.Message{
		Type: icmpType, Code: 0,
		Body: &icmp.Echo{
			ID:   p.id,
			Seq:  p.seq,
			Data: p.payload(),
		},
	}).Marshal(nil)
	if err != nil {
		log.Println(err)
	}

	wg.Add(1)
	go func(conn *icmp.PacketConn, dest net.Addr, b []byte) {
		defer wg.Done()
		for {
			if _, err := conn.WriteTo(bytes, dest); err != nil {
				log.Println(err)
				if neterr, ok := err.(*net.OpError); ok {
					if neterr.Err == syscall.ENOBUFS {
						continue
					}
				}
			}
			break
		}
	}(conn, p.addr, bytes)

	wg.Wait()
}

func (p *Ping) Ping(out chan string) {
	var (
		conn     *icmp.PacketConn
		rcvdChan chan *packet = make(chan *packet, 1)
	)

	if p.isV4Avail {
		if conn = p.listen("ip4:icmp"); conn == nil {
			return
		}
		defer conn.Close()
	}

	if p.isV6Avail {
		if conn = p.listen("ip6:ipv6-icmp"); conn == nil {
			return
		}
		defer conn.Close()
	}

	p.send(conn)
	p.recv(conn, rcvdChan)
	m := <-rcvdChan
	rm, err := icmp.ParseMessage(1, m.bytes)
	if err != nil {
		log.Println(err)
	}
	switch rm.Body.(type) {
	case *icmp.TimeExceeded:
		log.Println("time exceeded")
	case *icmp.PacketTooBig:
		log.Println("packet too big")
	case *icmp.DstUnreach:
		log.Println("unreachable")
	case *icmp.Echo:
		out <- fmt.Sprintf("%f ms", float64(time.Now().UnixNano()-getTimeStamp(m.bytes))/1000000)
	default:
		log.Println("error")
	}

}

func getTimeStamp(m []byte) int64 {
	var ts int64
	for i := uint(0); i < 8; i++ {
		ts += int64(m[uint(len(m))-8+i]) << (i * 8)
	}
	return ts
}

func (p *Ping) Start() {

}
