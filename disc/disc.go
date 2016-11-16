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

	"golang.org/x/net/icmp"
	"golang.org/x/net/ipv4"
	"golang.org/x/net/ipv6"

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
	OUI   map[string]string
	IsBSD bool
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
		IsBSD: IsBSD(),
		OUI:   make(map[string]string, 25000),
		SKey:  key,
	}
}

// WalkIP tries to walk through subnet as generator
func WalkIP(cidr string) chan string {
	c := make(chan string, 2048)
	go func() {
		defer close(c)
		ip, ipnet, err := net.ParseCIDR(cidr)
		if err != nil {
			println(err.Error())
			return
		}
		for ip := ip.Mask(ipnet.Mask); ipnet.Contains(ip); nextIP(ip) {
			select {
			case c <- ip.String():
			default:
				break
			}
		}
	}()
	return c
}

// PingLan tries to send a tiny UDP packet to all LAN hosts
func (a *disc) PingLan() {
	var (
		isV4 bool
		b    [16]byte
	)

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
				fd, err := syscall.Socket(syscall.AF_INET, syscall.SOCK_RAW, syscall.IPPROTO_ICMP)
				if err != nil {
					println(err.Error())
					return
				}
				syscall.SetsockoptInt(fd, 0x0, syscall.IP_TTL, 1)
				for ipStr := range WalkIP(addr.String()) {
					copy(b[:], net.ParseIP(ipStr).To4())
					addr := syscall.SockaddrInet4{
						Port: 33434,
						Addr: [4]byte{b[0], b[1], b[2], b[3]},
					}
					m, _ := (&icmp.Message{
						Type: ipv4.ICMPTypeEcho, Code: 0,
						Body: &icmp.Echo{
							ID: 2016, Seq: 1,
							Data: make([]byte, 52-28),
						},
					}).Marshal(nil)
					if err := syscall.Sendto(fd, m, 0, &addr); err != nil {
						println(err.Error())
					}
				}
				syscall.Close(fd)
			} else {
				fd, err := syscall.Socket(syscall.AF_INET6, syscall.SOCK_RAW, syscall.IPPROTO_ICMPV6)
				if err != nil {
					println(err.Error())
					return
				}
				syscall.SetsockoptInt(fd, syscall.IPPROTO_IPV6, syscall.IPV6_UNICAST_HOPS, 1)
				copy(b[:], net.ParseIP(fmt.Sprintf("ff02::1")).To16())
				addr := syscall.SockaddrInet6{
					Port:   33434,
					ZoneId: uint32(i.Index),
					Addr:   b,
				}
				m, _ := (&icmp.Message{
					Type: ipv6.ICMPTypeEchoRequest, Code: 0,
					Body: &icmp.Echo{
						ID: 2016, Seq: 1,
						Data: make([]byte, 52-48),
					},
				}).Marshal(nil)
				if err := syscall.Sendto(fd, m, 0, &addr); err != nil {
					println(err.Error())
				}
				syscall.Close(fd)
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

// StrTobyte16 converts string to 16 bytes
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
	if a.IsBSD {
		err1 := a.GetMACOSIPv6Neighbor()
		err2 := a.GetMACOSARPTable()
		if err1 != nil {
			return err1
		}
		if err2 != nil {
			return err2
		}
	} else {
		return a.GetLinuxARPTable()
	}

	return nil
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

		// incompleted arp request
		if fields[2] == "0x0" {
			continue
		}

		if len(hosts) == 0 {
			host = "NA"
		} else {
			host = hosts[0]
		}

		a.Table = append(a.Table, ARP{IP: fields[0], Host: host, MAC: fields[3], Interface: fields[5]})
	}
	return nil
}

// GetIPv6Neighbor gets existing NDP entries
func (a *disc) GetMACOSIPv6Neighbor() error {
	cmd := exec.Command("ndp", "-an")
	outBytes, err := cmd.Output()
	if err != nil {
		return err
	}
	out := strings.TrimSpace(string(outBytes))
	for _, l := range strings.Split(out, "\n") {
		fields := strings.Fields(l)

		if len(fields) < 1 || fields[0] == "Neighbor" || fields[3] == "permanent" || fields[3] == "expired" {
			continue
		}
		ipV6Address := strings.Split(fields[0], "%")
		//TODO: waiting for Go resolver to lookup ipv6 address to name, the current milestone is Go1.8
		a.Table = append(a.Table, ARP{IP: ipV6Address[0], Host: "NA", MAC: fields[1], Interface: fields[2]})
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

// GetOUILive gets oui info from iEEE
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
	println(counter, "host(s) has been found")
}

// IsBSD checks OS
func IsBSD() bool {
	if runtime.GOOS != "linux" {
		return true
	}
	return false
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
