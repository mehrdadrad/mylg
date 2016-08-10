// Package packet is a wrapper for GoPacket and sub packages
package packet

import (
	"fmt"
	"log"
	"net"
	"os"
	"os/signal"
	"strings"
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
	Payload string
	// info
	device string
}

// logWriter represents custom writer
type logWriter struct {
}

type Options struct {
	// no color
	nc bool
	// count
	c int
	// device list
	d bool
	// write pcap
	w string
}

var (
	snapLen     int32 = 6 * 1024
	promiscuous       = false
	err         error
	timeout     = 100 * time.Nanosecond
	handle      *pcap.Handle
	addrs       = make(map[string]struct{}, 20)
	ifName      string

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
		c:  cli.SetFlag(flag, "c", 1000000).(int),
		d:  cli.SetFlag(flag, "d", false).(bool),
		w:  cli.SetFlag(flag, "w", "").(string),
	}

	if options.d {
		printDev()
		return nil, nil
	}

	log.SetFlags(0)
	log.SetOutput(new(logWriter))

	return &Packet{
		device: cli.SetFlag(flag, "i", "").(string),
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

	// return first available interface and all ip addresses
	ifName, addrs = lookupDev()
	if p.device == "" {
		p.device = ifName
	}

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
	log.Printf("%s %s > %s",
		czStr("ARP      ", color.FgWhite, color.BgBlue),
		net.HardwareAddr(p.ARP.SourceHwAddress),
		net.HardwareAddr(p.ARP.DstHwAddress))
}

// PrintIPv4 prints IPv4 packets
func (p *Packet) PrintIPv4() {

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
}

// PrintIPv6 prints IPv6 packets
func (p *Packet) PrintIPv6() {

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
	return fmt.Printf("%s %s", time.Now().Format("15:04:05.000"), string(bytes))
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
		p.Payload = string(applicationLayer.Payload())
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
	host, _ := net.LookupAddr(ip.String())
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
		if i.Flags == 19 && ifName == "" {
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
		status  = "DOWN"
		columns []string
	)

	table := tablewriter.NewWriter(os.Stdout)
	table.SetHeader([]string{"Name", "MAC", "Status", "MTU", "Multicast", "Broadcast", "PointToPoint", "Loopback"})
	ifs, _ := net.Interfaces()
	for _, i := range ifs {
		if strings.Contains(i.Flags.String(), "up") {
			status = "UP"
		} else {
			status = "DOWN"
		}
		columns = append(columns, i.Name, i.HardwareAddr.String(), status, fmt.Sprintf("%d", i.MTU))
		for _, flag := range []string{"multicast", "broadcast", "pointtopoint", "loopback"} {
			if strings.Contains(i.Flags.String(), flag) {
				columns = append(columns, "\u2713")
			} else {
				columns = append(columns, "")
			}
		}
		table.Append(columns)
		columns = columns[:0]
	}
	table.Render()
}

func help() {
	fmt.Println(`
    usage:
          dump [-c count][-i interface][-w filename][-d][-nc]
    options:		  
          -c count       Stop after receiving count packets (default: 1M)
          -i interface   Listen on specified interface (default: first non-loopback)
          -w filename    Write packets to a pcap format file            
          -d             Print list of available interfaces 		  
          -nc            Shows dumps without color
    Example:
          dump tcp and port 443 -c 1000
          dump !udp
          dump -i eth0
          dump -w /tmp/mypcap		  
	`)
}
