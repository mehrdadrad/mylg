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
	"strings"
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
func NewPing(args string, cfg cli.Config) (*Ping, error) {
	var (
		err error
	)
	target, flag := cli.Flag(args)

	// show help
	if _, ok := flag["help"]; ok || len(target) < 3 {
		help(cfg)
		return nil, nil
	}

	p := Ping{
		id:        rand.Intn(0xffff),
		seq:       -1,
		pSize:     64,
		target:    target,
		isV4Avail: false,
		isV6Avail: false,
		isCIDR:    isCIDR(target),
		count:     cli.SetFlag(flag, "c", cfg.Ping.Count).(int),
		forceV4:   cli.SetFlag(flag, "4", false).(bool),
		forceV6:   cli.SetFlag(flag, "6", false).(bool),
		network:   "ip",
		source:    "",
		MaxRTT:    time.Second,
	}

	if !p.isCIDR {
		// resolve host
		ips, err := net.LookupIP(target)
		if err != nil {
			return nil, err
		}
		p.addrs = ips
		if err := p.SetIP(ips); err != nil {
			return nil, err
		}

	}
	// set timeout
	timeoutStr := cli.SetFlag(flag, "t", cfg.Ping.Timeout).(string)
	timeoutStr = NormalizeDuration(timeoutStr)
	if p.timeout, err = time.ParseDuration(timeoutStr); err != nil {
		return nil, fmt.Errorf("timeout options is not valid")
	}
	// set interval
	intervalStr := cli.SetFlag(flag, "i", cfg.Ping.Interval).(string)
	intervalStr = NormalizeDuration(intervalStr)
	if p.interval, err = time.ParseDuration(intervalStr); err != nil {
		return nil, fmt.Errorf("interval options is not valid")
	}

	return &p, nil
}

// Run loops the ping and print it out
func (p *Ping) Run() chan Response {
	var r = make(chan Response, 1)
	go func() {
		for n := 0; n < p.count; n++ {
			p.Ping(r)
			if n != p.count-1 {
				time.Sleep(p.interval)
			}
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
	if r := strings.Index(ip.String(), ":"); r != -1 {
		return true
	}
	return false
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
func (p *Ping) SetIP(ips []net.IP) error {
	for _, ip := range ips {
		if IsIPv4(ip) && !p.forceV6 {
			p.addr = &net.IPAddr{IP: ip}
			p.isV4Avail = true
			return nil
		} else if IsIPv6(ip) && !p.forceV4 {
			p.addr = &net.IPAddr{IP: ip}
			p.isV6Avail = true
			return nil
		}
	}
	return fmt.Errorf("there is not  A or AAAA record")
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
		println(err.Error())
	}

	wg.Add(1)
	go func(conn *icmp.PacketConn, dest net.Addr, b []byte) {
		defer wg.Done()
		for {
			if _, err := conn.WriteTo(bytes, dest); err != nil {
				println(err.Error())
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
		out <- Response{Error: rm.err, Sequence: p.seq}
		return
	}
	_, m, err := p.parseMessage(rm)
	if err != nil {
		out <- Response{Error: err, Sequence: p.seq}
		return
	}

	switch m.Body.(type) {
	case *icmp.TimeExceeded:
		out <- Response{Error: fmt.Errorf("time exceeded"), Sequence: p.seq}
	case *icmp.PacketTooBig:
		out <- Response{Error: fmt.Errorf("packet too big"), Sequence: p.seq}
	case *icmp.DstUnreach:
		out <- Response{Error: fmt.Errorf("destination unreachable"), Sequence: p.seq}
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
		out <- Response{Error: fmt.Errorf("ICMP error"), Sequence: p.seq}
	}
}

// PrintPretty prints out the result pretty format
func (p *Ping) PrintPretty(resp chan Response) {
	var (
		loop          = true
		sigCh         = make(chan os.Signal, 1)
		pFmt          = "%d bytes from %s icmp_seq=%d time=%f ms"
		eFmt          = "%s icmp_seq=%d"
		sFmt          = "%d packets transmitted,  %d packets received, %d%% packet loss\n"
		msg           string
		min, max, avg float64
		c             = map[string]int{"tx": 0, "err": 0, "pl": 0}
	)

	// capture interrupt w/ s channel
	signal.Notify(sigCh, os.Interrupt)
	defer signal.Stop(sigCh)

	fmt.Printf("PING %s (%s): %d data bytes\n", p.target, p.addr, p.pSize-8)
	for loop {
		select {
		case r, ok := <-resp:
			if !ok {
				loop = false
				break
			}
			c["tx"]++
			if r.Error != nil {
				c["err"]++
				msg = fmt.Sprintf(eFmt, r.Error.Error(), r.Sequence)
				println(msg)
				continue
			}

			min = Min(r.RTT, min)
			max = Max(r.RTT, max)
			avg = Avg(r.RTT, avg)

			msg = fmt.Sprintf(pFmt, r.Size, r.Addr, r.Sequence, r.RTT)
			println(msg)
		case <-sigCh:
			loop = false
		}
	}

	if c["tx"] == 0 {
		return
	}
	// packet loss
	c["pl"] = c["err"] * 100 / c["tx"]

	fmt.Printf("\n--- %s ping statistics ---\n", p.target)
	fmt.Printf(sFmt, c["tx"], c["tx"]-c["err"], c["pl"])

	if c["pl"] == 100 {
		return
	}

	fmt.Printf("round-trip min/avg/max = %.3f/%.3f/%.3f ms\n", min, avg, max)
}

// IsCIDR
func (p *Ping) IsCIDR() bool {
	return p.isCIDR
}

// Max handles maximum delay
func Max(x, y float64) float64 {
	if x > y {
		return x
	}
	return y
}

// Min handles minimum delay
func Min(x, y float64) float64 {
	if x < y || y == 0 {
		return x
	}
	return y
}

// Avg handles average delay
func Avg(x, y float64) float64 {
	if y == 0 {
		return x
	}
	return (x + y) / 2
}

// getTimeStamp
func getTimeStamp(m []byte) int64 {
	var ts int64
	for i := uint(0); i < 8; i++ {
		ts += int64(m[uint(len(m))-8+i]) << (i * 8)
	}
	return ts
}

// isCIDR
func isCIDR(s string) bool {
	if _, _, err := net.ParseCIDR(s); err != nil {
		return false
	}
	return true
}

// NormalizeDuration adds default unit (seconds) as needed
func NormalizeDuration(d string) string {
	if match, _ := regexp.MatchString(`^\d+\.{0,1}\d*$`, d); match {
		return d + "s"
	}
	return d
}

// help represents ping help
func help(cfg cli.Config) {
	fmt.Printf(`
    usage:
          ping IP address / domain name [options]
    options:
          -c count       Send 'count' requests (default: %d)
          -t timeout     Specify a timeout in format "ms", "s", "m" (default: %s)
          -i interval    Wait interval between sending each packet (default: %s)
          -4             Forces the ping command to use IPv4 (target should be hostname)
          -6             Forces the ping command to use IPv6 (target should be hostname)
    Example:
          ping 8.8.8.8
          ping 8.8.8.8 -c 10
          ping google.com -6
          ping mylg.io -i 5s
	`,
		cfg.Ping.Count,
		cfg.Ping.Timeout,
		cfg.Ping.Interval)
}
