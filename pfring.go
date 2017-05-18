package main

import (
	"fmt"
	"flag"
	"github.com/google/gopacket/pfring"
	"github.com/google/gopacket/layers"
	"github.com/google/gopacket"
	"time"
	"net"
)

const ttl = 32
const magic = 42

func main() {
	var device = flag.String("device", "eno1", "capture device")
	var caplen = flag.Int("caplen", 1024, "capture length")

	var src = flag.String("srcIp", "10.60.0.175", "source IPv4")
	var dst = flag.String("dstIp", "8.8.8.8", "dest IPv4")

	var srcM = flag.String("srcMAC", "aa:aa:aa:aa:aa:aa", "source MAC")
	var dstM = flag.String("dstMAC", "aa:aa:aa:aa:aa:aa", "gateway MAC")

	flag.Parse()

	srcIP4 := net.ParseIP(*src).To4()
	dstIP4 := net.ParseIP(*dst).To4()

	srcMAC, err := net.ParseMAC(*srcM)
	if err != nil {
		panic(err)
	}
	dstMAC, err := net.ParseMAC(*dstM)
	if err != nil {
		panic(err)
	}

	fmt.Printf("PF_RING ping timing demo: ping %s from %s\n", dstIP4, srcIP4)

	handle, err := pfring.NewRing(*device, uint32(*caplen), pfring.FlagPromisc)
	if err != nil {
		panic(err)
	}
	defer handle.Close()

	if err := handle.SetDirection(pfring.ReceiveOnly); err != nil {
		panic(err)
	}

	if err := handle.Enable(); err != nil {
		panic(err)
	}

	// Build ping packet in buf
	buf := gopacket.NewSerializeBuffer()
	opts := gopacket.SerializeOptions{
		FixLengths:       true,
		ComputeChecksums: true,
	}
	eth := layers.Ethernet{
		SrcMAC:       srcMAC,
		DstMAC:       dstMAC,
		EthernetType: layers.EthernetTypeIPv4,
	}
	ip4 := layers.IPv4{
		SrcIP:    srcIP4,
		DstIP:	  dstIP4,
		Version:  4,
		TTL:      ttl,
		Protocol: layers.IPProtocolICMPv4,
	}
	ping := layers.ICMPv4{
		TypeCode: layers.CreateICMPv4TypeCode(layers.ICMPv4TypeEchoRequest, 0),
		Id: magic,
	}
	if err := gopacket.SerializeLayers(buf, opts, &eth, &ip4, &ping); err != nil {
		panic(err)
	}

	// Set up parser for the inbound packet
	var deth layers.Ethernet
	var dip4 layers.IPv4
	var dicmp4 layers.ICMPv4
	parser := gopacket.NewDecodingLayerParser(layers.LayerTypeEthernet, &deth, &dip4, &dicmp4)
	decoded := []gopacket.LayerType{}
	flow := gopacket.NewFlow(layers.EndpointIPv4, dstIP4, srcIP4)
	data := make([]byte, *caplen)

	// Send ping 5 times
	for c := 0; c < 5; c++ {
		start := time.Now()
		handle.WritePacketData(buf.Bytes())

		for {
			if _, err = handle.ReadPacketDataTo(data); err != nil {
				panic(err)
			}

			err = parser.DecodeLayers(data, &decoded)

			// 1. no error and the packet has the layers we expect
			if err != nil || len(decoded) != 3 {
				continue // skip packets we are not interested in
			}

			// 2. it is part of the src / dst IP network flow we expect
			if dip4.NetworkFlow() != flow {
				continue
			}

			// 3. has the magic ICMP id we expected
			if dicmp4.Id != magic {
				continue
			}

			// Got it back
			break
		}

		// Dump timing
		end := time.Now()
		fmt.Printf("ping time=%s\n", end.Sub(start))
	}
}
