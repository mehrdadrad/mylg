// packet is a wrapper for GoPacket and sub packages
package packet

import (
	"fmt"
	"log"
	"net"
	"os"
	"os/signal"
	"time"

	"github.com/google/gopacket"
	"github.com/google/gopacket/layers"
	"github.com/google/gopacket/pcap"
)

// TCPFlags holds TCP flags
type TCPFlags struct {
	FIN, SYN, RST, PSH, ACK, URG, ECE, CWR, NS bool
}

// Packet holds all layers information
type Packet struct {
	EthType    layers.EthernetType
	SrcMAC     net.HardwareAddr
	DstMAC     net.HardwareAddr
	Src        net.IP
	Dst        net.IP
	Proto      layers.IPProtocol
	Flags      TCPFlags
	SrcPort    int
	DstPort    int
	Seq        uint32
	Ack        uint32
	Window     uint16
	Urgent     uint16
	Checksum   uint16
	Payload    string
	DataOffset uint8
}

var (
	device            = "en0"
	snapshotLen int32 = 1024
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

	go func() {
		handle, err = pcap.OpenLive(device, snapshotLen, promiscuous, timeout)
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
			case c <- GetPacketInfo(packet):
			}
		}
	}()
	return c
}

// PrintPretty prints out the captured data
// to the stdout
func (p *Packet) PrintPretty() {
	switch p.EthType {
	case layers.EthernetTypeIPv4:
		p.PrintIPv4()
	case layers.EthernetTypeIPv6:
		println("IPV6")
		p.PrintIPv4()
	case layers.EthernetTypeARP:
		println("ARP")
	default:
		println("unknown")

	}
}

// PrintIPv4 prints IPv4 packets
func (p *Packet) PrintIPv4() {
	switch {
	case p.Proto == layers.IPProtocolTCP:
		log.Printf("IP %s > %s , %s length: %d\n", p.Src, p.Dst, p.Proto, len(p.Payload))
	case p.Proto == layers.IPProtocolUDP:
		log.Printf("IP %s > %s , %s length: %d\n", p.Src, p.Dst, p.Proto, len(p.Payload))
	}
}

// GetPacketInfo -------
func GetPacketInfo(packet gopacket.Packet) *Packet {
	var p Packet
	// Ethernet
	ethernetLayer := packet.Layer(layers.LayerTypeEthernet)
	if ethernetLayer != nil {
		ethernetPacket, _ := ethernetLayer.(*layers.Ethernet)
		p.SrcMAC = ethernetPacket.SrcMAC
		p.DstMAC = ethernetPacket.DstMAC
		p.EthType = ethernetPacket.EthernetType
	}
	// IP Address V4
	ipLayer := packet.Layer(layers.LayerTypeIPv4)
	if ipLayer != nil {
		ip, _ := ipLayer.(*layers.IPv4)
		// IP layer variables:
		// Version (Either 4 or 6)
		// IHL (IP Header Length in 32-bit words)
		// TOS, Length, Id, Flags, FragOffset, TTL, Protocol (TCP?),
		// Checksum, SrcIP, DstIP
		p.Src = ip.SrcIP
		p.Dst = ip.DstIP
		p.Proto = ip.Protocol
	}

	// TCP
	tcpLayer := packet.Layer(layers.LayerTypeTCP)
	if tcpLayer != nil {
		tcp, _ := tcpLayer.(*layers.TCP)

		p.Seq = tcp.Seq
		p.Ack = tcp.Ack
		p.Window = tcp.Window
		p.Urgent = tcp.Urgent
		p.Checksum = tcp.Checksum
		p.DataOffset = tcp.DataOffset

		p.Flags.FIN = tcp.FIN
		p.Flags.SYN = tcp.SYN
		p.Flags.RST = tcp.RST
		p.Flags.PSH = tcp.PSH
		p.Flags.ACK = tcp.ACK
		p.Flags.URG = tcp.URG
		p.Flags.ECE = tcp.ECE
		p.Flags.CWR = tcp.CWR
		p.Flags.NS = tcp.NS

		p.SrcPort = int(tcp.SrcPort)
		p.DstPort = int(tcp.DstPort)
	} else {
		// UDP
		udpLayer := packet.Layer(layers.LayerTypeUDP)
		if udpLayer != nil {
			udp, _ := udpLayer.(*layers.UDP)
			p.SrcPort = int(udp.SrcPort)
			p.DstPort = int(udp.DstPort)
		}
	}
	// Iterate over all layers, printing out each layer type
	//fmt.Println("All packet layers:")
	//for _, layer := range packet.Layers() {
	//	fmt.Println("- ", layer.LayerType())
	//}

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
