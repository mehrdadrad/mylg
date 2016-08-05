// Package disc is a LAN discovery library
package disc

import (
	"bufio"
	"encoding/csv"
	"fmt"
	"io/ioutil"
	"net"
	"net/http"
	"os"
	"os/exec"
	"runtime"
	"strings"
	"syscall"
	"time"

	"github.com/olekukonko/tablewriter"

	"github.com/mehrdadrad/mylg/cli"
)

const (
	// IEEEOUI holds ieee oui csv file url
	IEEEOUI = "http://standards.ieee.org/develop/regauth/oui/oui.csv"
)

// ARP holds ARP information
type ARP struct {
	IP        string
	MAC       string
	Interface string
	Host      string
}

// disc holds all discovery information
type disc struct {
	Table []ARP
	IPs   []string
	OUI   map[string]string
	IsMac bool
	SKey  string
}

// New creates new discovery object
func New(args string) *disc {
	key, flag := cli.Flag(args)

	// help
	if _, ok := flag["help"]; ok {
		help()
		return nil
	}

	return &disc{
		IsMac: IsMac(),
		OUI:   make(map[string]string, 25000),
		SKey:  key,
	}
}

// WalkIP tries to salk through subnet as generator
func WalkIP(cidr string) chan string {
	c := make(chan string, 1)
	go func() {
		defer close(c)
		ip, ipnet, err := net.ParseCIDR(cidr)
		if err != nil {
			println(err.Error())
			return
		}
		for ip := ip.Mask(ipnet.Mask); ipnet.Contains(ip); nextIP(ip) {
			c <- ip.String()
		}
	}()
	return c
}

// PingLan tries to send a tiny UDP packet to all LAN hosts
func (a *disc) PingLan() {
	var isV4 bool
	fd, err := syscall.Socket(syscall.AF_INET, syscall.SOCK_DGRAM, syscall.IPPROTO_UDP)
	if err != nil {
		println(err.Error())
		return
	}
	defer syscall.Close(fd)
	// Set options
	syscall.SetsockoptInt(fd, 0x0, syscall.IP_TTL, 10)
	ifs, _ := net.Interfaces()
	for _, i := range ifs {
		addrs, _ := i.Addrs()
		if i.Flags != 19 {
			continue
		}
		// ip network(s) that assigned to interface(s)
		for _, addr := range addrs {
			if strings.IndexAny(addr.String(), "::") != -1 {
				isV4 = false
			} else {
				isV4 = true
			}
			if isV4 {
				a.IPs = append(a.IPs, addr.String())
				for ipStr := range WalkIP(addr.String()) {
					ip := net.ParseIP(ipStr).To4()
					addr := syscall.SockaddrInet4{
						Port: 33434,
						Addr: [4]byte{ip[0], ip[1], ip[2], ip[3]},
					}
					syscall.Sendto(fd, []byte{}, 0, &addr)
				}
			} else {
				//IPv6 doesn't support for the time being
				// ToDo
			}
		}
	}
}
func nextIP(ip net.IP) {
	for j := len(ip) - 1; j >= 0; j-- {
		ip[j]++
		if ip[j] > 0 {
			break
		}
	}
}

// StrToByte16 converts string to 16 bytes
func StrTobyte16(s string) [16]byte {
	var r [16]byte
	if len(s) > 16 {
		copy(r[:], s)
	} else {
		copy(r[16-len(s):], s)
	}
	return r
}

// GetARPTable gets ARP table
func (a *disc) GetARPTable() error {
	if a.IsMac {
		return a.GetMACOSARPTable()
	} else {
		return a.GetLinuxARPTable()
	}
	return fmt.Errorf("")
}

// GetLinuxARPTable gets Linux ARP table
func (a *disc) GetLinuxARPTable() error {
	var host string
	file, err := os.Open("/proc/net/arp")
	if err != nil {
		return err
	}
	s := bufio.NewScanner(file)
	s.Scan() // skip description
	for s.Scan() {
		fields := strings.Fields(s.Text())
		hosts, _ := net.LookupAddr(fields[0])

		if len(hosts) == 0 {
			host = "NA"
		} else {
			host = hosts[0]
		}

		a.Table = append(a.Table, ARP{IP: fields[0], Host: host, MAC: fields[3], Interface: fields[5]})
	}
	return nil
}

// GetMACOSARPTable gets Mac OS X ARP table
func (a *disc) GetMACOSARPTable() error {
	var host string
	cmd := exec.Command("arp", "-an")
	outBytes, err := cmd.Output()
	if err != nil {
		return err
	}
	out := strings.TrimSpace(string(outBytes))
	for _, l := range strings.Split(out, "\n") {
		fields := strings.Fields(l)
		if len(fields) < 1 {
			continue
		}
		if fields[3] != "(incomplete)" {
			fields[1] = strings.Trim(fields[1], ")")
			fields[1] = strings.Trim(fields[1], "(")
			hosts, _ := net.LookupAddr(fields[1])

			if len(hosts) == 0 {
				host = "NA"
			} else {
				host = hosts[0]
			}

			a.Table = append(a.Table, ARP{IP: fields[1], Host: host, MAC: fields[3], Interface: fields[5]})
		}
	}
	return nil
}

// LoadOUI
func (a *disc) LoadOUI() bool {
	if _, ok := cache("validate", nil); ok {
		if c, ok := cache("read", nil); ok {
			r := csv.NewReader(strings.NewReader(c))
			records, _ := r.ReadAll()
			for _, record := range records {
				if len(record) > 2 {
					a.OUI[record[1]] = record[2]
				}
			}
			return true
		}

	} else {
		b, err := GetOUILive()
		if err != nil {
			println(err.Error())
			return false
		}
		if c, ok := cache("write", b); ok {
			r := csv.NewReader(strings.NewReader(c))
			records, _ := r.ReadAll()
			for _, record := range records {
				if len(record) > 2 {
					a.OUI[record[1]] = record[2]
				}
			}
			return true
		}
	}
	return false
}

// GETOUI gets oui info from iEEE
func GetOUILive() ([]byte, error) {
	resp, err := http.Get(IEEEOUI)
	if err != nil {
		return []byte{}, fmt.Errorf("regauth.standards.ieee.org is unreachable (1)")
	}
	if resp.StatusCode != 200 {
		return []byte{}, fmt.Errorf("regauth.standards.ieee.org returns none 200 HTTP code")
	}
	defer resp.Body.Close()
	body, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		return []byte{}, fmt.Errorf("peeringdb.com is unreachable (2)  %s", err.Error())
	}
	return body, nil
}
func cache(r string, b []byte) (string, bool) {
	var (
		err error
		res string
	)
	switch r {
	case "write":
		err = ioutil.WriteFile("/tmp/mylg.disc", b, 0644)
		if err != nil {
			return "", false
		}
		res = string(b)
		return res, true
	case "read":
		b, err := ioutil.ReadFile("/tmp/mylg.disc")
		if err != nil {
			return "", false
		}
		res = string(b)
		return res, true
	case "validate":
		f, err := os.Stat("/tmp/mylg.disc")
		if err != nil {
			return "", false
		}
		d := time.Since(f.ModTime())
		if d.Hours() > 24*10 {
			return "", false
		}
	}

	return "", true
}

// PrintPretty prints ARP table
func (a *disc) PrintPretty() {
	var (
		orgName string
		counter = 0
	)
	table := tablewriter.NewWriter(os.Stdout)
	table.SetHeader([]string{"IP", "MAC", "Host", "Interface", "Organization Name"})
	for _, arp := range a.Table {

		// OUI
		if name, ok := a.OUI[strings.ToUpper(strings.Replace(arp.MAC, ":", "", -1))[:6]]; ok {
			if len(name) > 25 {
				orgName = name[:25] + "..."
			} else {
				orgName = name
			}
		} else {
			orgName = "NA"
		}

		// Search
		sHost := search(arp.Host, a.SKey)
		sARP := search(arp.IP, a.SKey)
		sMAC := search(arp.MAC, a.SKey)
		sOrg := search(orgName, a.SKey)

		if !sHost && !sARP && !sMAC && !sOrg {
			continue
		}

		table.Append([]string{arp.IP, arp.MAC, arp.Host, arp.Interface, orgName})
		counter++
	}
	table.Render()
	println(counter, "host has been found")
}

// IsMac checks OS
func IsMac() bool {
	if runtime.GOOS != "darwin" {
		return false
	}
	return true
}

// search tries to find key at data
func search(data, key string) bool {
	data = strings.ToLower(data)
	key = strings.ToLower(key)
	if strings.Contains(data, key) {
		return true
	}
	return false
}

// help shows disc help
func help() {
	fmt.Println(`
    Network LAN Discovery	
    usage:
          disc [search keyword]
    Example:
          disc
          disc 5c:a:5b:aa:4a:99
          disc apple
          disc 192.168.0.10		  
	`)
}
