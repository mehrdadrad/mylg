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

	"github.com/fatih/color"

	"github.com/mehrdadrad/mylg/cli"
)

// Packet holds all layers information
type Packet struct {
	// packet layers data
	Eth    *layers.Ethernet
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

var (
	snapLen     int32 = 6 * 1024
	promiscuous       = false
	err         error
	timeout     = 100 * time.Nanosecond
	handle      *pcap.Handle
	addrs       = make(map[string]struct{}, 20)
	ifName      string

	noColor bool
	filter  string
	count   int
)

// NewPacket creates an empty packet info
func NewPacket(args string) (*Packet, error) {
	var flag map[string]interface{}

	filter, flag = cli.Flag(args)

	// help
	if _, ok := flag["help"]; ok {
		help()
		return nil, fmt.Errorf("help")
	}

	noColor = cli.SetFlag(flag, "nc", false).(bool)
	count = cli.SetFlag(flag, "c", 1000000).(int)

	if err != nil {
		return nil, err
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
				c <- ParsePacketLayers(packet)
				if counter++; counter > count-1 {
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
		// todo
	default:
		// todo
	}
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

func (writer logWriter) Write(bytes []byte) (int, error) {
	return fmt.Printf("%s %s", time.Now().Format("15:04:05.000"), string(bytes))
}

// ParsePacketLayers decodes layers
func ParsePacketLayers(packet gopacket.Packet) *Packet {
	var p Packet
	// Ethernet
	ethernetLayer := packet.Layer(layers.LayerTypeEthernet)
	if ethernetLayer != nil {
		p.Eth, _ = ethernetLayer.(*layers.Ethernet)
	}

	// IP Address V4
	ipLayer := packet.Layer(layers.LayerTypeIPv4)
	if ipLayer != nil {
		p.IPv4, _ = ipLayer.(*layers.IPv4)
		p.SrcHost = lookup(p.IPv4.SrcIP)
		p.DstHost = lookup(p.IPv4.DstIP)
	} else {
		// IP Address V6
		ipLayer := packet.Layer(layers.LayerTypeIPv6)
		if ipLayer != nil {
			p.IPv6, _ = ipLayer.(*layers.IPv6)
			p.SrcHost = lookup(p.IPv6.SrcIP)
			p.DstHost = lookup(p.IPv6.DstIP)
		}
	}

	// TCP
	tcpLayer := packet.Layer(layers.LayerTypeTCP)
	if tcpLayer != nil {
		p.TCP, _ = tcpLayer.(*layers.TCP)
	} else {
		// UDP
		udpLayer := packet.Layer(layers.LayerTypeUDP)
		if udpLayer != nil {
			p.UDP, _ = udpLayer.(*layers.UDP)
		}
	}

	// ICMPv4
	icmpLayer := packet.Layer(layers.LayerTypeICMPv4)
	if icmpLayer != nil {
		p.ICMPv4, _ = icmpLayer.(*layers.ICMPv4)
	} else {
		// ICMPv6
		icmpv6Layer := packet.Layer(layers.LayerTypeICMPv6)
		if icmpv6Layer != nil {
			p.ICMPv6, _ = icmpv6Layer.(*layers.ICMPv6)
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
	if _, ok := addrs[ip.String()]; ok && !noColor {
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
	if !noColor {
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

func help() {
	fmt.Println(`
    usage:
          dump [-c count][-nc]
    options:		  
          -c count       Stop after receiving count packets (default: 1M)
          -nc            Shows dumps without color
    Example:
          dump tcp and port 443 -c 1000
	`)
}
