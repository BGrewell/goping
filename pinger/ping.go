package pinger

import (
	"encoding/binary"
	"fmt"
	"golang.org/x/net/icmp"
	"golang.org/x/net/ipv4"
	"net"
	"os"
	"time"
)

var (
	dests map[string]*net.IPAddr
)

func init() {
	// Automatically called
	dests = make(map[string]*net.IPAddr)
}

func getListener() (*icmp.PacketConn, error) {

	conn, err := icmp.ListenPacket("ip4:icmp", "0.0.0.0")
	if err != nil {
		return nil, err
	}
	return conn, err
}

func Ping(target string, seq uint16, timeout time.Duration) (dest *net.IPAddr, sendtime int64, rtt time.Duration, err error) {

	// Get a listener so we can hear the ping reply
	conn, err := getListener()
	if err != nil {
		return nil, 0, 0, err
	}
	defer conn.Close()

	// Resolve DNS if used
	if val, ok := dests[target]; ok {
		dest = val
	} else {
		dest, err = net.ResolveIPAddr("ip4", target)
		if err != nil {
			return nil, 0, 0, err
		}
	}

	// Create the ICMP packet
	payload := []byte("abcdefghijklmnopqrstuvwxyz0123456789")
	packet := icmp.Message{
		Type: ipv4.ICMPTypeEcho,
		Code: 0,
		Body: &icmp.Echo{
			ID:   os.Getpid() & 0xffff,
			Seq:  int(seq),
			Data: payload,
		},
	}

	// Get the packet bytes
	packetBytes, err := packet.Marshal(nil)
	if err != nil {
		return dest, 0, 0, err
	}

	// Send
	start := time.Now()
	written, err := conn.WriteTo(packetBytes, dest)
	if err != nil {
		return dest, start.UnixNano(), 0, err
	} else if written != len(packetBytes) {
		return dest, start.UnixNano(), 0, fmt.Errorf("incomplete write, wanted: %v got: %v", len(packetBytes), written)
	}

	// Wait for reply
	replyBytes := make([]byte, 1500)


	replySeq := uint16(0)
	var read int
	var peer net.Addr
	for replySeq != seq {
		err = conn.SetReadDeadline(time.Now().Add(timeout))
		if err != nil {
			return dest, start.UnixNano(),0, err
		}
		read, peer, err = conn.ReadFrom(replyBytes)
		replySeq = binary.BigEndian.Uint16(replyBytes[6:8])
		if err != nil {
			return dest, start.UnixNano(), 0, err
		}
	}

	rtt = time.Since(start)

	// should never hit this code
	if seq != replySeq {
		fmt.Printf("Reply sequence number wrong. Expected: %v Got: %v", seq, replySeq)
	}
	// Parse reply
	reply, err := icmp.ParseMessage(1, replyBytes[:read])
	if err != nil {
		return dest, start.UnixNano(), 0, err
	}

	switch reply.Type {
	case ipv4.ICMPTypeEchoReply:
		return dest, start.UnixNano(), rtt, nil
	default:
		return dest, start.UnixNano(),0, fmt.Errorf("[Error] Type: %v Peer: %v", reply.Type, peer)
	}
}
