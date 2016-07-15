// Package scan TCP ports
package scan

import (
	"fmt"
	"net"
	"regexp"
	"strconv"
	"sync"
	"time"
)

// Scan represents the scan parameters
type Scan struct {
	minPort int
	maxPort int
	target  string
}

// NewScan creats scan object
func NewScan(args string) (Scan, error) {
	var (
		scan Scan
		err  error
	)
	re := regexp.MustCompile(`([^\s]+)\s+(\d+)\s+(\d+)`)
	f := re.FindStringSubmatch(args)
	if len(f) == 4 {
		scan.target = f[1]
		scan.minPort, err = strconv.Atoi(f[2])
		scan.maxPort, err = strconv.Atoi(f[3])
		if err != nil {
			return scan, err
		}
	} else {
		scan.target = args
		scan.minPort = 1
		scan.maxPort = 100
	}
	if !scan.isCIDR() {
		ipAddr, err := net.ResolveIPAddr("ip4", scan.target)
		if err != nil {
			return scan, err
		}
		scan.target = ipAddr.String()
	}
	return scan, nil
}

// isCIDR checks the target if it's CIDR
func (s Scan) isCIDR() bool {
	_, _, err := net.ParseCIDR(s.target)
	if err != nil {
		return false
	}
	return true
}

// Run tries to scan wide range ports (TCP)
func (s Scan) Run() {
	if !s.isCIDR() {
		host(s.target, s.minPort, s.maxPort)
	}
}

// host tries to scan a single host
func host(ipAddr string, minPort, maxPort int) {
	var (
		wg      sync.WaitGroup
		tStart  = time.Now()
		counter int
	)
	for i := minPort; i <= maxPort; i++ {
		wg.Add(1)
		go func(i int) {
			conn, err := net.DialTimeout("tcp", fmt.Sprintf("%s:%d", ipAddr, i), 1*time.Second)
			if err != nil {
				wg.Done()
				return
			}
			counter++
			fmt.Printf("%d/tcp open\n", i)
			wg.Done()
			conn.Close()
		}(i)
		time.Sleep(8 * time.Millisecond)
	}
	wg.Wait()
	if counter == 0 {
		println("there isn't any opened port")
	} else {
		elapsed := fmt.Sprintf("%.3f seconds", time.Since(tStart).Seconds())
		println("Scan done:", counter, "opened port found in", elapsed)
	}
}
