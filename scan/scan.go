// Package scan
package scan

import (
	"fmt"
	"net"
	"regexp"
	"strconv"
	"sync"
	"time"
)

// Run tries to scan wide range ports (TCP)
func Run(args string) {
	var (
		wg      sync.WaitGroup
		host    string
		minPort int
		maxPort int
		err     error
		counter = 0
	)
	re := regexp.MustCompile(`([^\s]+)\s+(\d+)\s+(\d+)`)
	f := re.FindStringSubmatch(args)
	if len(f) == 4 {
		host = f[1]
		minPort, err = strconv.Atoi(f[2])
		maxPort, err = strconv.Atoi(f[3])
	} else {
		host = args
		minPort = 1
		maxPort = 100
	}

	ipAddr, err := net.ResolveIPAddr("ip4", host)
	if err != nil {
		println(err.Error())
		return
	}
	for i := minPort; i <= maxPort; i++ {
		wg.Add(1)
		go func(i int) {
			conn, err := net.DialTimeout("tcp", fmt.Sprintf("%s:%d", ipAddr, i), 1*time.Second)
			if err != nil {
				wg.Done()
				return
			}
			counter++
			println("OPEN", i, "TCP")
			wg.Done()
			conn.Close()
		}(i)
		time.Sleep(10 * time.Millisecond)
	}
	wg.Wait()
	if counter == 0 {
		println("there isn't any opened port")
	}

}
