// Package scan TCP ports
package scan

import (
	"fmt"
	"net"
	"regexp"
	"strconv"
	"sync"
	"time"

	"github.com/mehrdadrad/mylg/cli"
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
		flag map[string]interface{}
		err  error
	)

	args, flag = cli.Flag(args)
	// help
	if _, ok := flag["help"]; ok {
		help()
		return scan, fmt.Errorf("")
	}

	pRange := cli.SetFlag(flag, "p", "1-500").(string)

	re := regexp.MustCompile(`(\d+)-(\d+)`)
	f := re.FindStringSubmatch(pRange)
	if len(f) == 3 {
		scan.target = args
		scan.minPort, err = strconv.Atoi(f[1])
		scan.maxPort, err = strconv.Atoi(f[2])
		if err != nil {
			return scan, err
		}
	}
	if !scan.IsCIDR() {
		ipAddr, err := net.ResolveIPAddr("ip4", scan.target)
		if err != nil {
			return scan, err
		}
		scan.target = ipAddr.String()
	}
	return scan, nil
}

// IsCIDR checks the target if it's CIDR
func (s Scan) IsCIDR() bool {
	_, _, err := net.ParseCIDR(s.target)
	if err != nil {
		return false
	}
	return true
}

// Run tries to scan wide range ports (TCP)
func (s Scan) Run() {
	if !s.IsCIDR() {
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
			err = conn.Close()
			if err != nil {
				println(err.Error())
			}
		}(i)
		time.Sleep(8 * time.Millisecond)
	}
	wg.Wait()
	if counter == 0 {
		println("there isn't any opened port")
	} else {
		elapsed := fmt.Sprintf("%.3f seconds", time.Since(tStart).Seconds())
		println("Scan done:", counter, "opened port(s) found in", elapsed)
	}
}

// help represents guide to user
func help() {
	println(`
    usage:
          scan ip/host [-p portrange]
    example:
          scan www.google.com -p 1-500
	`)
}
