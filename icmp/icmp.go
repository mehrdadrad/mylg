// Package icmp provides icmp enhanced methods based on the golang icmp package
package icmp

import (
	"errors"
	"fmt"
	"golang.org/x/net/icmp"
	"golang.org/x/net/ipv4"
	"golang.org/x/net/ipv6"
	"log"
	"math/rand"
	"net"
	"os"
	"os/signal"
	"regexp"
	"sync"
	"syscall"
	"time"

	"github.com/mehrdadrad/mylg/cli"
)

// IANA ICMP
const (
	ProtocolIPv4ICMP = 1  // IANA ICMP for IPv4
	ProtocolIPv6ICMP = 58 // IANA ICMP for IPv6
)

// packet represents ping packet
type packet struct {
	bytes []byte
	addr  net.Addr
	err   error
}

// Ping represents ping request
type Ping struct {
	m         icmp.Message
	id        int
	seq       int
	pSize     int
	count     int
	addr      *net.IPAddr
	isV4Avail bool
	isV6Avail bool
	network   string
	source    string
	timeout   time.Duration
	interval  time.Duration
	MaxRTT    time.Duration
	mu        sync.RWMutex
}

// Response represent ping response
type Response struct {
	RTT      float64
	Size     int
	Sequence int
	Addr     string
	Timeout  bool
	Error    error
}

// NewPing creates a new ping object
func NewPing(args string) (*Ping, error) {
	var (
		timeout  time.Duration
		interval time.Duration
		err      error
	)
	target, flag := cli.Flag(args)

	// show help
	if _, ok := flag["help"]; ok || len(target) < 3 {
		help()
		return nil, nil
	}
	// resolve host
	ra, err := net.ResolveIPAddr("ip", target)
	if err != nil {
		return nil, err
	}
	// timeout
	timeoutStr := cli.SetFlag(flag, "t", "2s").(string)
	timeoutStr = NormalizeDuration(timeoutStr)
	if timeout, err = time.ParseDuration(timeoutStr); err != nil {
		return nil, fmt.Errorf("timeout options is not valid")
	}
	// interval
	intervalStr := cli.SetFlag(flag, "i", "1s").(string)
	intervalStr = NormalizeDuration(intervalStr)
	if interval, err = time.ParseDuration(intervalStr); err != nil {
		return nil, fmt.Errorf("interval options is not valid")
	}

	p := Ping{
		id:        rand.Intn(0xffff),
		seq:       -1,
		pSize:     64,
		count:     cli.SetFlag(flag, "c", 4).(int),
		timeout:   timeout,
		interval:  interval,
		isV4Avail: false,
		isV6Avail: false,
		network:   "ip",
		source:    "",
		MaxRTT:    time.Second,
	}

	p.SetIP(ra.String())
	return &p, nil
}

// Run loops the ping and print it out
func (p *Ping) Run() chan Response {
	var r = make(chan Response, 1)
	go func() {
		for n := 0; n < p.count; n++ {
			p.Ping(r)
			time.Sleep(p.interval)
		}
		close(r)
	}()
	return r
}

// IsIPv4 returns true if ip version is v4
func IsIPv4(ip net.IP) bool {
	return len(ip.To4()) == net.IPv4len
}

// IsIPv6 returns true if ip version is v6
func IsIPv6(ip net.IP) bool {
	return len(ip) == net.IPv6len
}

func (p *Ping) parseMessage(m *packet) (*ipv4.Header, *icmp.Message, error) {
	var proto = ProtocolIPv4ICMP
	if p.isV6Avail {
		proto = ProtocolIPv6ICMP
	}
	msg, err := icmp.ParseMessage(proto, m.bytes)
	if err != nil {
		return nil, nil, err
	}
	bytes, _ := msg.Body.Marshal(msg.Type.Protocol())
	h, err := icmp.ParseIPv4Header(bytes)
	return h, msg, err
}

// SetIP set ip address
func (p *Ping) SetIP(ipAddr string) {
	ip := net.ParseIP(ipAddr)
	p.mu.Lock()
	defer p.mu.Unlock()
	p.addr = &net.IPAddr{IP: ip}
	if IsIPv4(ip) {
		p.isV4Avail = true
	} else {
		p.isV6Avail = true
	}
}

// DelIP removes ip adrress
func (p *Ping) DelIP(ipAddr string) {
	//todo
}

// PacketSize set packet size
func (p *Ping) PacketSize(s int) {
	p.pSize = s
}

// listen starts to listen incoming icmp
func (p *Ping) listen(network string) *icmp.PacketConn {
	c, err := icmp.ListenPacket(network, p.source)
	if err != nil {
		log.Fatalf("listen err, %s", err)
	}
	return c
}

// recv reads icmp message
func (p *Ping) recv(conn *icmp.PacketConn, rcvdChan chan<- *packet) {
	var err error
	bytes := make([]byte, 1500)
	conn.SetReadDeadline(time.Now().Add(p.timeout))
	n, dest, err := conn.ReadFrom(bytes)
	if err != nil {
		if neterr, ok := err.(*net.OpError); ok {
			if neterr.Timeout() {
				err = errors.New("Request timeout")
			}
		}
	}

	bytes = bytes[:n]
	select {
	case rcvdChan <- &packet{bytes: bytes, addr: dest, err: err}:
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
	if IsIPv4(p.addr.IP) {
		icmpType = ipv4.ICMPTypeEcho
	} else if IsIPv6(p.addr.IP) {
		icmpType = ipv6.ICMPTypeEchoRequest
	}

	p.seq++
	bytes, err := (&icmp.Message{
		Type: icmpType, Code: 0,
		Body: &icmp.Echo{
			ID:   p.id,
			Seq:  p.seq,
			Data: p.payload(),
		},
	}).Marshal(nil)
	if err != nil {
		println(err)
	}

	wg.Add(1)
	go func(conn *icmp.PacketConn, dest net.Addr, b []byte) {
		defer wg.Done()
		for {
			if _, err := conn.WriteTo(bytes, dest); err != nil {
				println(err)
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

// Ping tries to send and receive a packet
func (p *Ping) Ping(out chan Response) {
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
	rm := <-rcvdChan

	if rm.err != nil {
		out <- Response{Error: rm.err}
		return
	}
	_, m, err := p.parseMessage(rm)
	if err != nil {
		out <- Response{Error: err}
		return
	}

	switch m.Body.(type) {
	case *icmp.TimeExceeded:
		log.Println("time exceeded")
	case *icmp.PacketTooBig:
		log.Println("packet too big")
	case *icmp.DstUnreach:
		log.Println("unreachable")
	case *icmp.Echo:
		rtt := float64(time.Now().UnixNano()-getTimeStamp(rm.bytes)) / 1000000
		out <- Response{
			Size:     len(rm.bytes),
			Addr:     rm.addr.String(),
			RTT:      rtt,
			Sequence: p.seq,
			Error:    nil,
		}
	default:
		log.Println("error")
	}

}

// PrintPretty prints out the result pretty format
func (p *Ping) PrintPretty() {
	var (
		loop  = true
		sigCh = make(chan os.Signal, 1)
		pFmt  = "%d bytes from %s icmp_seq=%d time=%f ms"
		resp  = p.Run()
	)

	// capture interrupt w/ s channel
	signal.Notify(sigCh, os.Interrupt)
	defer signal.Stop(sigCh)

	for loop {
		select {
		case r, ok := <-resp:
			if !ok {
				loop = false
				break
			}
			if r.Error != nil {
				println(r.Error.Error())
				continue
			}
			msg := fmt.Sprintf(pFmt, r.Size, r.Addr, r.Sequence, r.RTT)
			println(msg)
		case <-sigCh:
			loop = false
		}
	}
}

// getTimeStamp
func getTimeStamp(m []byte) int64 {
	var ts int64
	for i := uint(0); i < 8; i++ {
		ts += int64(m[uint(len(m))-8+i]) << (i * 8)
	}
	return ts
}

// NormalizeDuration adds default unit (seconds) as needed
func NormalizeDuration(d string) string {
	if match, _ := regexp.MatchString(`^\d+\.{0,1}\d*$`, d); match {
		return d + "s"
	}
	return d
}

// help represents ping help
func help() {
	fmt.Println(`
    usage:
          ping IP address / domain name
    options:		  
          -c count       Send 'count' requests (default: 4)
          -t timeout     Specifiy a timeout in format "ms", "s", "m" (default: 2s) 
          -i interval    Specifiy a interval in format "ms", "s", "m" (default: 1s) 
    Example:
          ping 8.8.8.8
          ping 8.8.8.8 -c 10
          ping mylg.io
	`)
}
