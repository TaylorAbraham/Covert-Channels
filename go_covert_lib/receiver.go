// Author Michael Dysart
package main

import (
	"./ipv4TCP"
	"flag"
	"fmt"
	"log"
	"net"
	"time"
)

func main() {
	fmt.Println("Covert Channel Receiver!")

	// Parse flags for destination and bounce IPs
	fa := flag.String("fa", "127.0.0.1", "The friend IP")
	oa := flag.String("oa", "127.0.0.1", "The origin IP")
	fp := flag.Int("fp", 8082, "The friend port")
	op := flag.Int("op", 8081, "The origin port")

	flag.Parse()

	fIP := net.ParseIP(*fa)
	oIP := net.ParseIP(*oa)

	if fIP == nil {
		log.Fatal("Invalid Friend IP")
	}
	if oIP == nil {
		log.Fatal("Invalid Origin IP")
	}

	conf := ipv4TCP.Config{
		FriendPort: uint16(*fp),
		OriginPort: uint16(*op),
		Delimiter:  ipv4TCP.Protocol,

		ReadTimeout : time.Second * 10,
	}

	copy(conf.FriendIP[:], fIP.To4())
	copy(conf.OriginIP[:], oIP.To4())

	for {

		ch, _ := ipv4TCP.MakeChannel(conf)

		fmt.Println("Waiting for message")

		var data [1024]byte

		n, err := ch.Receive(data[:], nil)
		if err != nil {
			fmt.Println("error: " + err.Error())
		} else {
			fmt.Println("Msg Received: " + string(data[:n]))
		}
	}
}
