package icmp

import (
	"fmt"
	"net"
	"os"
	"os/signal"
	"sort"
	"strings"
	"syscall"
	"time"

	"github.com/mehrdadrad/mylg/cli"
	"github.com/mehrdadrad/mylg/ripe"
)

const (
	// ICMP Time Exceeded type number
	ICMPTypeTimeExceeded int = 11
	// Default TX timeout
	DefaultTXTimeout int64 = 1000
	// Default RX timeout
	DefaultRXTimeout int64 = 2000
)

// Trace represents trace properties
type Trace struct {
	src     string
	host    string
	ips     []net.IP
	ttl     int
	fd      int
	timeout int64
	resolve bool
	ripe    bool
	maxTTL  int
}

// HopResp represents hop's response
type HopResp struct {
	hop     string
	ip      string
	elapsed float64
	last    bool
	whois   Whois
}

// Whois represents prefix info from RIPE
type Whois struct {
	holder string
	asn    float64
}

// MHopResp represents multi hop's responses
type MHopResp []HopResp

// NewTrace creates new trace object
func NewTrace(args string) (*Trace, error) {
	target, flag := cli.Flag(args)

	// show help
	if _, ok := flag["help"]; ok || len(target) < 3 {
		helpTrace()
		return nil, nil
	}
	ips, err := net.LookupIP(target)
	if err != nil {
		return nil, err
	}
	return &Trace{
		host:    target,
		ips:     ips,
		timeout: DefaultRXTimeout,
		resolve: cli.SetFlag(flag, "n", true).(bool),
		ripe:    cli.SetFlag(flag, "nr", true).(bool),
		maxTTL:  cli.SetFlag(flag, "m", 30).(int),
	}, nil
}

func (h MHopResp) Len() int           { return len(h) }
func (h MHopResp) Swap(i, j int)      { h[i], h[j] = h[j], h[i] }
func (h MHopResp) Less(i, j int) bool { return len(h[i].ip) > len(h[j].ip) }

// SetTTL set the IP packat time to live
func (i *Trace) SetTTL(ttl int) {
	i.ttl = ttl
}

// Send tries to send an UDP packet
func (i *Trace) Send() error {
	fd, err := syscall.Socket(syscall.AF_INET, syscall.SOCK_DGRAM, syscall.IPPROTO_UDP)
	if err != nil {
		println(err.Error())
	}
	defer syscall.Close(fd)
	// Set options
	syscall.SetsockoptInt(fd, 0x0, syscall.IP_TTL, i.ttl)

	if IsIPv4(i.ips[0]) {
		var b [4]byte
		copy(b[:], i.ips[0].To4())
		addr := syscall.SockaddrInet4{
			Port: 33434,
			Addr: b,
		}
		if err := syscall.Sendto(fd, []byte{0x0}, 0, &addr); err != nil {
			return err
		}
	} else if IsIPv6(i.ips[0]) {
		var b [16]byte
		copy(b[:], i.ips[0].To16())
		addr := syscall.SockaddrInet6{
			Port: 33434,
			Addr: b,
		}
		if err := syscall.Sendto(fd, []byte{0x0}, 0, &addr); err != nil {
			return err
		}
	}
	return nil
}

// SetReadDeadLine sets rx timeout
func (i *Trace) SetReadDeadLine() error {
	tv := syscall.NsecToTimeval(1e6 * i.timeout)
	return syscall.SetsockoptTimeval(i.fd, syscall.SOL_SOCKET, syscall.SO_RCVTIMEO, &tv)
}

// SetWriteDeadLine sets tx timeout
func (i *Trace) SetWriteDeadLine() error {
	tv := syscall.NsecToTimeval(1e6 * DefaultTXTimeout)
	return syscall.SetsockoptTimeval(i.fd, syscall.SOL_SOCKET, syscall.SO_SNDTIMEO, &tv)
}

// SetDeadLine sets tx/rx timeout
func (i *Trace) SetDeadLine() error {
	err := i.SetReadDeadLine()
	if err != nil {
		return err
	}
	err = i.SetWriteDeadLine()
	if err != nil {
		return err
	}
	return nil
}

// Bind starts to listen for ICMP reply
func (i *Trace) Bind() {
	var err error
	i.fd, err = syscall.Socket(syscall.AF_INET, syscall.SOCK_RAW, syscall.IPPROTO_ICMP)
	if err != nil {
		println("e2", err.Error())
	}
	err = i.SetDeadLine()
	if err != nil {
		println(err.Error())
	}
	// TODO: IPv6
	b := net.ParseIP("0.0.0.0").To4()
	addr := syscall.SockaddrInet4{
		Port: 0,
		Addr: [4]byte{b[0], b[1], b[2], b[3]},
	}

	if err := syscall.Bind(i.fd, &addr); err != nil {
		println("e3", err.Error())
	}
}

// Recv gets the replied icmp packet
func (i *Trace) Recv(fd int) (int, int, string) {
	var typ, code int
	buf := make([]byte, 512)
	n, from, err := syscall.Recvfrom(fd, buf, 0)
	if err == nil {
		buf = buf[:n]
		typ = int(buf[20])  // ICMP Type
		code = int(buf[21]) // ICMP Code
	}
	if typ != 0 {
		b := from.(*syscall.SockaddrInet4).Addr
		return typ, code, fmt.Sprintf("%v.%v.%v.%v", b[0], b[1], b[2], b[3])
	}
	return typ, code, ""
}

// Done close the socket
func (i *Trace) Done() {
	syscall.Close(i.fd)
}

// NextHop pings the specific hop by set TTL
func (i *Trace) NextHop(fd, hop int) HopResp {
	var (
		r    HopResp
		name []string
	)
	i.SetTTL(hop)
	ts := time.Now().UnixNano()
	err := i.Send()
	if err != nil {
		println("e5", err.Error())
	}
	t, _, ip := i.Recv(fd)
	elapsed := time.Now().UnixNano() - ts
	if t == ICMPTypeTimeExceeded || ip == i.ips[0].String() {
		if i.resolve {
			name, _ = net.LookupAddr(ip)
		}
		if len(name) > 0 {
			r = HopResp{name[0], ip, float64(elapsed) / 1e6, false, Whois{}}
		} else {
			r = HopResp{"", ip, float64(elapsed) / 1e6, false, Whois{}}
		}
	}
	// readed to the target
	for _, h := range i.ips {
		if ip == h.String() {
			r.last = true
			break
		}
	}
	return r
}

// Run provides trace based on the other methods
func (i *Trace) Run(retry int) chan []HopResp {
	var (
		c = make(chan []HopResp, 1)
		r []HopResp
	)
	i.Bind()
	go func() {
		for h := 1; h <= i.maxTTL; h++ {
			for n := 0; n < retry; n++ {
				r = append(r, i.NextHop(i.fd, h))
			}
			if i.ripe {
				appendWhois(r[:])
			}
			c <- r
			if r[0].last || r[1].last || r[2].last {
				break
			}

			r = r[:0]
		}
		close(c)
		i.Done()
	}()
	return c
}

// PrintPretty prints out trace result
func (i *Trace) PrintPretty() {
	var (
		loop    = true
		counter int
		sigCh   = make(chan os.Signal, 1)
		resp    = i.Run(3)
	)

	signal.Notify(sigCh, os.Interrupt)
	defer signal.Stop(sigCh)

	// header
	fmt.Printf("trace route to %s (%s), %d hops max\n", i.host, i.ips[0], i.maxTTL)

	for loop {
		select {
		case r, ok := <-resp:
			if !ok {
				loop = false
				break
			}
			counter++
			sort.Sort(MHopResp(r))
			// there is not any load balancing and there is at least a timeout
			if r[0].ip != r[1].ip && (r[1].elapsed == 0 || r[2].elapsed == 0) {
				fmt.Printf("%-2d %s", counter, fmtHops(r, 1))
				continue
			}
			// there is not any load balancing and there is at least a timeout
			if r[1].ip != r[2].ip && (r[0].elapsed == 0 || r[1].elapsed == 0) {
				fmt.Printf("%-2d %s", counter, fmtHops(r, 1))
				continue
			}
			// there is not any load balancing and there is at least a timeout
			if r[0].ip == r[1].ip && r[0].elapsed != 0 && r[2].elapsed == 0 {
				fmt.Printf("%-2d %s %s", counter, fmtHops(r[0:2], 0), fmtHops(r[2:3], 1))
				continue
			}

			// load balance between three routes
			if r[0].ip != r[1].ip && r[1].ip != r[2].ip {
				fmt.Printf("%-2d %s   %s   %s", counter, fmtHops(r[0:1], 1), fmtHops(r[1:2], 1), fmtHops(r[2:3], 1))
				continue
			}
			// load balance between two routes
			if r[0].ip == r[1].ip && r[1].ip != r[2].ip {
				fmt.Printf("%-2d %s   %s", counter, fmtHops(r[0:2], 1), fmtHops(r[2:3], 1))
				continue
			}
			// load balance between two routes
			if r[0].ip != r[1].ip && r[1].ip == r[2].ip {
				fmt.Printf("%-2d %s   %s", counter, fmtHops(r[0:1], 1), fmtHops(r[1:3], 1))
				continue
			}
			// there is not any load balancing
			if r[0].ip == r[1].ip && r[1].ip == r[2].ip {
				fmt.Printf("%-2d %s", counter, fmtHops(r, 1))
			}
			//fmt.Printf("%#v\n", r)
		case <-sigCh:
			loop = false
		}
	}
}

func fmtHops(m []HopResp, newLine int) string {
	var (
		timeout = false
		msg     string
	)
	for _, r := range m {
		if (msg == "" || timeout) && r.hop != "" {
			if r.whois.asn != 0 {
				msg += fmt.Sprintf("%s (%s) [ASN %.0f/%s] ", r.hop, r.ip, r.whois.asn, strings.Fields(r.whois.holder)[0])
			} else {
				msg += fmt.Sprintf("%s (%s) ", r.hop, r.ip)
			}
		}
		if (msg == "" || timeout) && r.hop == "" && r.elapsed != 0 {
			if r.whois.asn != 0 {
				msg += fmt.Sprintf("%s [ASN %.0f/%s] ", r.ip, r.whois.asn, strings.Fields(r.whois.holder)[0])
			} else {
				msg += fmt.Sprintf("%s ", r.ip)
			}
		}
		if r.elapsed != 0 {
			msg += fmt.Sprintf("%.3f ms ", r.elapsed)
			timeout = false
		} else {
			msg += "* "
			timeout = true
		}
	}
	if newLine == 1 {
		msg += "\n"
	}
	return msg
}

// appendWhois adds whois info to response if available
func appendWhois(R []HopResp) {
	var ips = make(map[string]Whois, 3)
	for _, r := range R {
		ips[r.ip] = Whois{}
	}
	for ip, _ := range ips {
		if ip == "" {
			continue
		}
		w, err := whois(ip + "/24")
		if err != nil {
			continue
		}
		ips[ip] = w
	}
	for i, _ := range R {
		R[i].whois = ips[R[i].ip]
	}
}

// whois returns prefix whois info from RIPE
func whois(ip string) (Whois, error) {
	var resp Whois
	r := new(ripe.Prefix)
	r.Set(ip)
	r.GetData()
	data, ok := r.Data["data"].(map[string]interface{})
	if !ok {
		return Whois{}, fmt.Errorf("data not available")
	}
	asns := data["asns"].([]interface{})
	for _, h := range asns {
		resp.holder = h.(map[string]interface{})["holder"].(string)
		resp.asn = h.(map[string]interface{})["asn"].(float64)
	}
	return resp, nil
}

func helpTrace() {
	fmt.Println(`
    usage:
          trace IP address / domain name [options]
    options:
          -n             Do not try to map IP addresses to host names
          -nr            Do not try to map IP addresses to ASN,Holder (RIPE NCC)
          -m MAX_TTL     Specifies the maximum number of hops
    Example:
          trace 8.8.8.8
	`)

}
