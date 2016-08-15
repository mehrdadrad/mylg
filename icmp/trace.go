package icmp

import (
	"fmt"
	"net"
	"os"
	"os/signal"
	"sort"
	"syscall"
	"time"

	"github.com/mehrdadrad/mylg/cli"
)

// Trace represents trace properties
type Trace struct {
	src     string
	host    string
	ips     []net.IP
	ttl     int
	timeout int64
}

type HopResp struct {
	hop     string
	ip      string
	elapsed float64
	last    bool
}

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
		timeout: 5000,
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

// Bind starts to listen for ICMP reply
func (i *Trace) Bind() int {
	fd, err := syscall.Socket(syscall.AF_INET, syscall.SOCK_RAW, syscall.IPPROTO_ICMP)
	if err != nil {
		println("e2", err.Error())
	}
	tv := syscall.NsecToTimeval(1e6 * i.timeout)
	err = syscall.SetsockoptTimeval(fd, syscall.SOL_SOCKET, syscall.SO_RCVTIMEO, &tv)
	if err != nil {
		println(err.Error())
	}
	b := net.ParseIP("0.0.0.0").To4()
	addr := syscall.SockaddrInet4{
		Port: 0,
		Addr: [4]byte{b[0], b[1], b[2], b[3]},
	}

	if err := syscall.Bind(fd, &addr); err != nil {
		println("e3", err.Error())
	}
	return fd
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
func (i *Trace) Done(fd int) {
	syscall.Close(fd)
}

// NextHop pings the specific hop by set TTL
func (i *Trace) NextHop(fd, hop int) HopResp {
	var (
		r HopResp
	)
	i.SetTTL(hop)
	ts := time.Now().UnixNano()
	err := i.Send()
	if err != nil {
		println("e5", err.Error())
	}
	t, c, ip := i.Recv(fd)
	elapsed := time.Now().UnixNano() - ts
	if t != 0 {
		name, _ := net.LookupAddr(ip)
		if len(name) > 0 {
			r = HopResp{name[0], ip, float64(elapsed) / 1e6, false}
		} else {
			r = HopResp{"", ip, float64(elapsed) / 1e6, false}
		}
	}
	if c == 3 {
		r.last = true
	}
	return r
}

// Run provides trace based on the other methods
func (i *Trace) Run(retry int) chan []HopResp {
	var (
		c = make(chan []HopResp, 1)
		r []HopResp
	)
	fd := i.Bind()
	go func() {
		for h := 1; h < 30; h++ {
			for n := 0; n < retry; n++ {
				r = append(r, i.NextHop(fd, h))
			}
			c <- r
			if r[0].last {
				close(c)
				i.Done(fd)
				break
			}
			r = r[:0]
		}
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
	fmt.Printf("trace route to %s (%s), 30 hops max\n", i.host, i.ips[0])

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
			// load balance between three routes
			if r[0].ip != r[1].ip && r[0].ip != r[2].ip && r[1].ip != r[2].ip {
				fmt.Printf("%-2d %s %s %s", counter, fmtHops(r[0:1], 0), fmtHops(r[1:2], 0), fmtHops(r[2:3], 1))
				continue
			}
			// load balance between two routes
			if r[0].ip == r[1].ip && r[1].ip != r[2].ip {
				fmt.Printf("%-2d %s %s", counter, fmtHops(r[0:2], 0), fmtHops(r[2:3], 1))
				continue
			}
			// load balance between two routes
			if r[0].ip != r[1].ip && r[1].ip == r[2].ip {
				fmt.Printf("%-2d %s %s", counter, fmtHops(r[0:1], 0), fmtHops(r[1:3], 1))
				continue
			}
			// there is not any load balancing
			if r[0].ip == r[1].ip && r[1].ip == r[2].ip {
				fmt.Printf("%-2d %s", counter, fmtHops(r, 1))
			}

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
			msg += fmt.Sprintf("%s (%s) ", r.hop, r.ip)
		}
		if (msg == "" || timeout) && r.hop == "" {
			msg += fmt.Sprintf("%s ", r.ip)
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

func helpTrace() {
	fmt.Println(`
    usage:
          trace IP address / domain name
    options:

    Example:
          trace 8.8.8.8
	`)

}
