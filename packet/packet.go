// Package packet is a wrapper for GoPacket and sub packages
package packet

import (
	"encoding/hex"
	"fmt"
	"log"
	"net"
	"os"
	"os/signal"
	"strings"
	"sync"
	"time"

	"github.com/google/gopacket"
	"github.com/google/gopacket/layers"
	"github.com/google/gopacket/pcap"
	"github.com/google/gopacket/pcapgo"

	"github.com/fatih/color"
	"github.com/olekukonko/tablewriter"

	"github.com/mehrdadrad/mylg/cli"
)

// Packet holds all layers information
type Packet struct {
	// packet layers data
	Eth    *layers.Ethernet
	ARP    *layers.ARP
	IPv4   *layers.IPv4
	IPv6   *layers.IPv6
	TCP    *layers.TCP
	UDP    *layers.UDP
	ICMPv4 *layers.ICMPv4
	ICMPv6 *layers.ICMPv6

	SrcHost []string
	DstHost []string
	Payload []byte
	// info
	device string
}

// logWriter represents custom writer
type logWriter struct {
}

// Options represents dump options
type Options struct {
	// no color
	nc bool
	// no resolve dns
	n bool
	// count
	c int
	// device list
	d bool
	// write pcap
	w string
	// print w/o timestamp
	t bool
	// dump payload
	x bool
	// search at payload
	s string
}

type LookUpCache struct {
	rec map[string]string
	sync.RWMutex
}

var (
	snapLen     int32 = 6 * 1024
	promiscuous       = false
	err         error
	timeout     = 100 * time.Nanosecond
	handle      *pcap.Handle
	addrs       = make(map[string]struct{}, 20)
	dev         string
	luCache     LookUpCache

	options Options
	filter  string
)

// NewPacket creates an empty packet info
func NewPacket(args string) (*Packet, error) {
	var flag map[string]interface{}

	filter, flag = cli.Flag(args)

	// help
	if _, ok := flag["help"]; ok {
		help()
		return nil, nil
	}

	options = Options{
		nc: cli.SetFlag(flag, "nc", false).(bool),
		n:  cli.SetFlag(flag, "n", false).(bool),
		c:  cli.SetFlag(flag, "c", 1000000).(int),
		d:  cli.SetFlag(flag, "d", false).(bool),
		x:  cli.SetFlag(flag, "x", false).(bool),
		t:  cli.SetFlag(flag, "t", false).(bool),
		w:  cli.SetFlag(flag, "w", "").(string),
		s:  cli.SetFlag(flag, "s", "").(string),
	}

	if options.d {
		printDev()
		return nil, nil
	}

	log.SetFlags(0)
	log.SetOutput(new(logWriter))
	luCache.rec = make(map[string]string)

	// return first available interface and all ip addresses
	dev, addrs = lookupDev()

	return &Packet{
		device: cli.SetFlag(flag, "i", dev).(string),
	}, nil
}

// Open is a loop over packets
func (p *Packet) Open() chan *Packet {
	var (
		c    = make(chan *Packet, 1)
		s    = make(chan os.Signal, 1)
		w    *pcapgo.Writer
		loop = true
	)
	// capture interrupt w/ s channel
	signal.Notify(s, os.Interrupt)

	go func() {
		var counter int
		defer signal.Stop(s)
		defer close(s)
		defer close(c)

		// write to pcap if needed
		if options.w != "" {
			f, _ := os.Create(options.w)
			w = pcapgo.NewWriter(f)
			w.WriteFileHeader(uint32(snapLen), layers.LinkTypeEthernet)
			defer f.Close()
		}

		handle, err = pcap.OpenLive(p.device, snapLen, promiscuous, timeout)
		if err != nil {
			log.Println(err.Error())
			return
		}
		if err := handle.SetBPFFilter(filter); err != nil {
			log.Println(err.Error())
			return
		}

		defer handle.Close()
		packetSource := gopacket.NewPacketSource(handle, handle.LinkType())
		for loop {
			select {
			case packet := <-packetSource.Packets():
				// write to pcap if needed
				if options.w != "" {
					w.WritePacket(packet.Metadata().CaptureInfo, packet.Data())
				}
				c <- ParsePacketLayers(packet)
				if counter++; counter > options.c-1 {
					loop = false
				}
			case <-s:
				loop = false
			}
		}
	}()
	return c
}

// Banner prints out info that related to
// packet capturing
func (p *Packet) Banner() string {
	return fmt.Sprintf("Interface: %s, capture size: %d bytes",
		p.device,
		snapLen,
	)
}

// PrintPretty prints out the captured data
// to the stdout
func (p *Packet) PrintPretty() {
	switch p.Eth.EthernetType {
	case layers.EthernetTypeIPv4:
		p.PrintIPv4()
	case layers.EthernetTypeIPv6:
		p.PrintIPv6()
	case layers.EthernetTypeARP:
		p.PrintARP()
	default:
		// todo
	}
}

// PrintARP prints ARP header
func (p *Packet) PrintARP() {
	var (
		operation = "ARP/Reply"
		bgColor   = color.BgBlue
		srcAddr   = p.ARP.SourceProtAddress
		dstAddr   = p.ARP.DstProtAddress
		dstIP     net.IP
		srcIP     net.IP
	)

	// stop printing
	if p.isSilent() {
		return
	}

	switch p.ARP.Protocol {
	case layers.EthernetTypeIPv4:
		srcIP = net.IPv4(srcAddr[0], srcAddr[1], srcAddr[2], srcAddr[3])
		dstIP = net.IPv4(dstAddr[0], dstAddr[1], dstAddr[2], dstAddr[3])
	case layers.EthernetTypeIPv6:
		// todo
	}

	if p.ARP.Operation == layers.ARPRequest {
		operation = "ARP/Req. "
		bgColor = color.BgHiBlue
		log.Printf("%s %s who-has %s tell %s (%s)",
			czStr(operation, color.FgWhite, bgColor),
			p.ARP.Protocol,
			dstIP.String(),
			srcIP.String(),
			net.HardwareAddr(p.ARP.SourceHwAddress))
	} else {
		log.Printf("%s %s %s is-at %s",
			czStr(operation, color.FgWhite, bgColor),
			p.ARP.Protocol,
			srcIP.String(),
			net.HardwareAddr(p.ARP.SourceHwAddress))
	}
}

// PrintIPv4 prints IPv4 packets
func (p *Packet) PrintIPv4() {
	// stop printing
	if p.isSilent() {
		return
	}

	src := czIP(p.IPv4.SrcIP, p.SrcHost, color.Bold)
	dst := czIP(p.IPv4.DstIP, p.DstHost, color.Bold)

	switch {
	case p.IPv4.Protocol == layers.IPProtocolTCP:
		log.Printf("%s %s:%s > %s:%s [%s], win %d, len: %d\n",
			czStr("IPv4/TCP ", color.FgBlack, color.BgWhite),
			src, p.TCP.SrcPort, dst, p.TCP.DstPort,
			czStr(p.flagsString(), color.Bold),
			p.TCP.Window, len(p.Payload))
	case p.IPv4.Protocol == layers.IPProtocolUDP:
		log.Printf("%s %s:%s > %s:%s , len: %d\n",
			czStr("IPv4/UDP ", color.FgBlack, color.BgCyan),
			src, p.UDP.SrcPort, dst, p.UDP.DstPort, len(p.Payload))
	case p.IPv4.Protocol == layers.IPProtocolICMPv4:
		log.Printf("%s %s > %s: %s id %d, seq %d, len: %d\n",
			czStr("IPv4/ICMP", color.FgBlack, color.BgYellow),
			src, dst, p.ICMPv4.TypeCode.String(), p.ICMPv4.Id,
			p.ICMPv4.Seq, len(p.Payload))
	}

	// dump payload in hex format
	if options.x {
		fmt.Println(hex.Dump(p.Payload))
	}
}

// PrintIPv6 prints IPv6 packets
func (p *Packet) PrintIPv6() {
	// stop printing
	if p.isSilent() {
		return
	}

	src := czIP(p.IPv6.SrcIP, p.SrcHost, color.Bold)
	dst := czIP(p.IPv6.DstIP, p.DstHost, color.Bold)

	switch {
	case p.IPv6.NextHeader == layers.IPProtocolTCP:
		log.Printf("%s %s:%s > %s:%s, len: %d\n",
			czStr("IPv6/TCP ", color.FgBlack, color.BgHiWhite),
			src, p.TCP.SrcPort, dst, p.TCP.DstPort,
			len(p.Payload))
	case p.IPv6.NextHeader == layers.IPProtocolUDP:
		log.Printf("%s %s:%s > %s:%s, len: %d\n",
			czStr("IPv6/UDP ", color.FgBlack, color.BgHiCyan),
			src, p.UDP.SrcPort, dst, p.UDP.DstPort, len(p.Payload))
	case p.IPv6.NextHeader == layers.IPProtocolICMPv6:
		log.Printf("%s %s > %s: %s, len: %d\n",
			czStr("IPv6/ICMP", color.FgBlack, color.BgYellow),
			src, dst, p.ICMPv6.TypeCode.String(), len(p.Payload))
	}

	// dump payload in hex format
	if options.x {
		fmt.Println(hex.Dump(p.Payload))
	}
}
func (p *Packet) isSilent() bool {
	if options.s == "" {
		return false
	}
	payload := strings.ToLower(string(p.Payload))
	keyword := strings.ToLower(options.s)
	if !strings.Contains(payload, keyword) {
		return true
	}
	return false
}

// flags returns flags string except ack
func (p *Packet) flagsString() string {
	var (
		r     []string
		flags = []bool{p.TCP.FIN, p.TCP.SYN, p.TCP.RST, p.TCP.PSH, p.TCP.URG, p.TCP.ECE, p.TCP.NS}
		sign  = "FSRPUECN"
	)
	for i, flag := range flags {
		if flag {
			r = append(r, string(sign[i]))
		}
	}
	r = append(r, ".")
	return strings.Join(r, "")
}

func (writer logWriter) Write(bytes []byte) (int, error) {
	if !options.t {
		return fmt.Printf("%s %s", time.Now().Format("15:04:05.000"), string(bytes))
	}
	return fmt.Printf("%s %s", time.Now().Format(""), string(bytes))
}

// ParsePacketLayers decodes layers (Lazy Decoding)
func ParsePacketLayers(packet gopacket.Packet) *Packet {
	var p Packet
	for _, l := range packet.Layers() {
		switch l.LayerType() {
		case layers.LayerTypeEthernet:
			p.Eth, _ = l.(*layers.Ethernet)
		case layers.LayerTypeARP:
			p.ARP, _ = l.(*layers.ARP)
		case layers.LayerTypeIPv4:
			p.IPv4, _ = l.(*layers.IPv4)
			p.SrcHost = lookup(p.IPv4.SrcIP)
			p.DstHost = lookup(p.IPv4.DstIP)
		case layers.LayerTypeIPv6:
			p.IPv6, _ = l.(*layers.IPv6)
			p.SrcHost = lookup(p.IPv6.SrcIP)
			p.DstHost = lookup(p.IPv6.DstIP)
		case layers.LayerTypeICMPv4:
			p.ICMPv4, _ = l.(*layers.ICMPv4)
		case layers.LayerTypeICMPv6:
			p.ICMPv6, _ = l.(*layers.ICMPv6)
		case layers.LayerTypeTCP:
			p.TCP, _ = l.(*layers.TCP)
		case layers.LayerTypeUDP:
			p.UDP, _ = l.(*layers.UDP)
		}
	}

	// Application
	applicationLayer := packet.ApplicationLayer()
	if applicationLayer != nil {
		p.Payload = applicationLayer.Payload()
	}

	// Check for errors
	if err := packet.ErrorLayer(); err != nil {
		//fmt.Println("Error decoding some part of the packet:", err)
		// todo
	}
	return &p
}

// czIP makes colorize IP/Host
func czIP(ip net.IP, host []string, attr ...color.Attribute) string {
	var (
		src string
	)
	if _, ok := addrs[ip.String()]; ok && !options.nc {
		if len(host) > 0 {
			src = czStr(host[0], attr...)
		} else {
			src = czStr(ip.String(), attr...)
		}
	} else {
		if len(host) > 0 {
			src = host[0]
		} else {
			src = ip.String()
		}
	}
	return src
}

// czStr makes colorize string
func czStr(i string, attr ...color.Attribute) string {
	c := color.New(attr...).SprintfFunc()
	if !options.nc {
		return c(i)
	}
	return i
}

func lookup(ip net.IP) []string {
	// don't resolve
	if options.n {
		return []string{ip.String()}
	}
	// dns cache
	if r, ok := luCache.rec[ip.String()]; ok {
		return []string{r}
	}

	var (
		lu   = make(chan []string, 1)
		host []string
	)

	go func() {
		defer func() {
			recover()
		}()
		host, _ := net.LookupAddr(ip.String())
		lu <- host
		luCache.Lock()
		luCache.rec[ip.String()] = host[0]
		luCache.Unlock()
		// clean up cache
		if len(luCache.rec) > 1e3 {
			c := 1
			luCache.Lock()
			for _, k := range luCache.rec {
				delete(luCache.rec, k)
				c++
				if c > 100 {
					break
				}
			}
			luCache.Unlock()
		}
	}()

	select {
	case host = <-lu:
	case <-time.Tick(time.Duration(10) * time.Millisecond):
		host = []string{ip.String()}
	}

	close(lu)

	return host
}

func lookupDev() (string, map[string]struct{}) {
	var (
		ips    = make(map[string]struct{}, 20)
		ifName = ""
	)
	ifs, _ := net.Interfaces()
	for _, i := range ifs {
		addrs, _ := i.Addrs()
		if len(addrs) > 0 && i.Flags == 19 && ifName == "" {
			ifName = i.Name
		}
		for _, addr := range addrs {
			ips[strings.Split(addr.String(), "/")[0]] = struct{}{}
		}
	}
	return ifName, ips
}

func printDev() {
	var (
		status   = "DOWN"
		columns  []string
		addrsStr []string
	)

	table := tablewriter.NewWriter(os.Stdout)
	table.SetHeader([]string{"Name", "MAC", "Status", "MTU", "IP Addresses", "Multicast", "Broadcast", "PointToPoint", "Loopback"})

	ifs, _ := net.Interfaces()
	for _, i := range ifs {
		if strings.Contains(i.Flags.String(), "up") {
			status = "UP"
		} else {
			status = "DOWN"
		}
		addrs, _ := i.Addrs()
		for _, addr := range addrs {
			addrsStr = append(addrsStr, addr.String())
		}
		columns = append(columns, i.Name, i.HardwareAddr.String(), status, fmt.Sprintf("%d", i.MTU), strings.Join(addrsStr, "\n"))
		for _, flag := range []string{"multicast", "broadcast", "pointtopoint", "loopback"} {
			if strings.Contains(i.Flags.String(), flag) {
				columns = append(columns, "\u2713")
			} else {
				columns = append(columns, "")
			}
		}
		table.Append(columns)
		columns = columns[:0]
		addrsStr = addrsStr[:0]
	}
	table.Render()
}

func help() {
	fmt.Println(`
    usage:
          dump [filter expression] [options]
          * The expression consists of one or more primitives (Berkeley Packet Filter (BPF) syntax)
    options:
          -c count       Stop after receiving count packets (default: 1M)
          -i interface   Listen on specified interface (default: first non-loopback)
          -w filename    Write packets to a pcap format file
          -d             Print list of available interfaces
          -t             Print without timestamp on each dump line.
          -x             Dump payload in hex format
          -s keyword     Search keyword at payload
          -n             Don't convert host addresses to names
          -nc            Shows dumps without color
    Example:
          dump tcp and port 443 -c 1000
          dump !udp
          dump -i eth0
          dump -w /tmp/mypcap
          dump tcp -s verisign -x
  `)
}
