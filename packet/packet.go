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
)

// Packet holds all layers information
type Packet struct {
	Eth     *layers.Ethernet
	IPv4    *layers.IPv4
	IPv6    *layers.IPv6
	TCP     *layers.TCP
	UDP     *layers.UDP
	SrcHost []string
	DstHost []string
	Payload string
}

var (
	device            = "en0"
	snapLen     int32 = 1024
	promiscuous       = false
	err         error
	timeout     = 100 * time.Nanosecond
	handle      *pcap.Handle
)

// NewPacket creates an empty packet info
func NewPacket() *Packet {
	return &Packet{}
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
	getIFAddrs()
	go func() {
		handle, err = pcap.OpenLive(device, snapLen, promiscuous, timeout)
		if err != nil {
			log.Fatal(err)
		}
		if err := handle.SetBPFFilter(""); err != nil {
			log.Fatal(err)
		}

		defer handle.Close()
		defer close(s)
		defer close(c)

		packetSource := gopacket.NewPacketSource(handle, handle.LinkType())
		for loop {
			packet, err := packetSource.NextPacket()
			if err != nil {
				continue
			}
			select {
			case <-s:
				loop = false
				signal.Stop(s)
			case c <- ParsePacketLayers(packet):
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
		println("IPV6")
		p.PrintIPv6()
	case layers.EthernetTypeARP:
		println("ARP")
	default:
		println("unknown")

	}
}

// PrintIPv4 prints IPv4 packets
func (p *Packet) PrintIPv4() {
	var src, dst string

	if len(p.SrcHost) > 0 {
		src = p.SrcHost[0]
	} else {
		src = p.IPv4.SrcIP.String()
	}
	if len(p.DstHost) > 0 {
		dst = p.DstHost[0]
	} else {
		dst = p.IPv4.DstIP.String()
	}

	switch {
	case p.IPv4.Protocol == layers.IPProtocolTCP:
		log.Printf("IP4/%s %s:%s > %s:%s [%s], len: %d\n",
			p.IPv4.Protocol, src, p.TCP.SrcPort, dst, p.TCP.DstPort, p.flags(), len(p.Payload))
	case p.IPv4.Protocol == layers.IPProtocolUDP:
		log.Printf("IP4/%s %s:%s > %s:%s , len: %d\n",
			p.IPv4.Protocol, src, p.UDP.SrcPort, dst, p.UDP.DstPort, len(p.Payload))
	}
}

// flags returns flags string except ack
func (p *Packet) flags() string {
	var f []string
	if p.TCP.FIN {
		f = append(f, "F")
	}
	if p.TCP.SYN {
		f = append(f, "S")
	}
	if p.TCP.RST {
		f = append(f, "R")
	}
	if p.TCP.PSH {
		f = append(f, "P")
	}
	if p.TCP.URG {
		f = append(f, "U")
	}
	if p.TCP.ECE {
		f = append(f, "E")
	}
	if p.TCP.CWR {
		f = append(f, "C")
	}
	if p.TCP.NS {
		f = append(f, "N")
	}
	f = append(f, ".")
	return strings.Join(f, "")
}

// PrintIPv6 prints IPv6 packets
func (p *Packet) PrintIPv6() {
	var src, dst string

	if len(p.SrcHost) > 0 {
		src = p.SrcHost[0]
	} else {
		src = p.IPv6.SrcIP.String()
	}
	if len(p.DstHost) > 0 {
		dst = p.DstHost[0]
	} else {
		dst = p.IPv6.DstIP.String()
	}

	switch {
	case p.IPv6.NextHeader == layers.IPProtocolTCP:
		log.Printf("IP6/%s %s:%s > %s:%s , len: %d\n",
			p.IPv6.NextHeader, src, p.TCP.SrcPort, dst, p.TCP.DstPort, len(p.Payload))
	case p.IPv6.NextHeader == layers.IPProtocolUDP:
		log.Printf("IP6/%s %s:%s > %s:%s , len: %d\n",
			p.IPv6.NextHeader, src, p.UDP.SrcPort, dst, p.UDP.DstPort, len(p.Payload))
	}
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

	// Application
	applicationLayer := packet.ApplicationLayer()
	if applicationLayer != nil {
		p.Payload = string(applicationLayer.Payload())
	}

	// Check for errors
	if err := packet.ErrorLayer(); err != nil {
		fmt.Println("Error decoding some part of the packet:", err)
	}
	return &p
}

func lookup(ip net.IP) []string {
	host, _ := net.LookupAddr(ip.String())
	return host
}

func getIFAddrs() map[string]struct{} {
	var r = make(map[string]struct{}, 20)

	ifs, _ := net.Interfaces()
	for _, i := range ifs {
		addrs, _ := i.Addrs()
		for _, addr := range addrs {
			r[addr.String()] = struct{}{}
		}
	}
	return r
}
