package icmp

import (
	"fmt"
	"net"
	"syscall"
	"time"
)

// Trace represents trace properties
type Trace struct {
	src     string
	dest    string
	ttl     int
	timeout int64
}

// Init set the basic parameters
func (i *Trace) Init(dest string, timeout int64) {
	i.dest = dest
	i.timeout = timeout
}

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

	b := net.ParseIP(i.dest).To4()
	addr := syscall.SockaddrInet4{
		Port: 33434,
		Addr: [4]byte{b[0], b[1], b[2], b[3]},
	}
	if err := syscall.Sendto(fd, []byte{0x0}, 0, &addr); err != nil {
		return err
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

// Run provides trace based on the other methods
func (i *Trace) Run(ip string) {
	var (
		trace Trace
		done  = false
	)
	ipAddr, err := net.ResolveIPAddr("ip4", ip)
	if err != nil {
		println("error: can not resolve")
		return
	}
	fmt.Printf("Traceroute to %s (%s), 30 hops max\n", ip, ipAddr.String())
	trace.Init(ipAddr.String(), 5000)
	fd := trace.Bind()
	for i := 1; i < 30; i++ {
		fmt.Printf("%d  ", i)
		for b := 0; b < 1; b++ {
			trace.SetTTL(i)
			ts := time.Now().UnixNano()
			err := trace.Send()
			if err != nil {
				println("e5", err.Error())
			}
			t, c, ip := trace.Recv(fd)
			elapsed := time.Now().UnixNano() - ts
			if t == 0 {
				fmt.Printf("*  ")
			} else {
				name, _ := net.LookupAddr(ip)
				if len(name) > 0 {
					fmt.Printf("%s (%s) %.3f ms ", name[0], ip, float64(elapsed)/1e6)
				} else {
					fmt.Printf("%s %.3f ms ", ip, float64(elapsed)/1e6)
				}
			}
			if c == 3 {
				done = true
			}
		}
		fmt.Printf("\n")
		if done {
			break
		}
	}
	trace.Done(fd)
}
