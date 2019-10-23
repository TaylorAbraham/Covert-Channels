// Author Michael Dysart
package main

import (
	"./ipv4TCP"
	"bufio"
	"flag"
	"fmt"
	"log"
	"net"
	"os"
)

func main() {
	fmt.Println("Covert Channel Sender!")

	// Parse flags for destination and bounce IPs
	fa := flag.String("fa", "127.0.0.1", "The friend IP")
	oa := flag.String("oa", "127.0.0.1", "The origin IP")
	fp := flag.Int("fp", 8081, "The friend port")
	op := flag.Int("op", 8082, "The origin port")

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
	}

	copy(conf.FriendIP[:], fIP.To4())
	copy(conf.OriginIP[:], oIP.To4())

	ch, _ := ipv4TCP.MakeChannel(conf)

	scanner := bufio.NewScanner(os.Stdin)

	for {
		fmt.Println("Write your message")

		scanner.Scan()
		text := scanner.Text()
		fmt.Printf("%d bytes read\n", len(text)+1)
		data := []byte(text + "\n")
		if len(data) > 1024 {
			data = data[:1024]
		}

		_, err := ch.Send(data, nil)
		if err != nil {
			fmt.Println("error: " + err.Error())
		} else {
			fmt.Println("Msg sent")
		}
	}
}
