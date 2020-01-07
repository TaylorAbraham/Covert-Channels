package main

import (
  "../controller/channel/tcp"
  "log"
)

func main() {

  cf := tcp.Config{
    FriendIP : [4]byte{127,0,0,1},
    OriginIP : [4]byte{127,0,0,1},
    FriendReceivePort : 1025,
    OriginReceivePort : 1027,
  }

  ch, err := tcp.MakeChannel(cf)
  if err != nil {
    log.Fatal(err.Error())
  }

  var buf [1024]byte

  defer ch.Close()

  for {
    n, err := ch.Receive(buf[:], nil)
    if err != nil {
      log.Println("Error : " + err.Error())
    } else {
      log.Println("Message : " + string(buf[:n]))
    }
  }
}
