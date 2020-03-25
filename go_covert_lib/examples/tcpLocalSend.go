package main

import (
	"../controller/channel/tcp"
	"bufio"
	"log"
	"os"
)

func main() {

	cf := tcp.Config{
		FriendIP:          [4]byte{127, 0, 0, 1},
		OriginIP:          [4]byte{127, 0, 0, 1},
		FriendReceivePort: 1027,
		OriginReceivePort: 1025,
	}

	ch, err := tcp.MakeChannel(cf)
	if err != nil {
		log.Fatal(err.Error())
	}

	scanner := bufio.NewScanner(os.Stdin)

	defer ch.Close()

	for {
		log.Println("Input your message")
		scanner.Scan()
		text := scanner.Text()

		buf := []byte(text)

		_, err := ch.Send(buf, nil)
		if err != nil {
			log.Println("Error : " + err.Error())
		} else {
			log.Println("Message : " + string(buf[:]))
		}
	}
}
