package controller

import (
  "github.com/gorilla/websocket"
  "sync"
  "./channel"
  "./channel/ipv4TCP"
  "./processor"
  "./processor/none"
)

// The go json library has this really convenient feature
// where if given a json string and a structure,
// it will only decode the values with keys in both json string
// and the structure
// This allows for selecting unmarshalling i.e. unmarshal once
// to get the type, and then unmarshall a second time to get the
// data. This protects against having to always use the same
// struct for communication
type command struct {
  OpCode string
}

type messageType struct {
  OpCode string
  Message string
}

type configType struct {
  OpCode string
  ChannelType string
  ProcessorType string
}

type configData struct {
  OpCode string
  ChannelType string
  ProcessorType string
  Processor processorData
  Channel   channelData
}

type channelData struct {
  Ipv4TCP ipv4TCP.ConfigClient
}

type processorData struct {
  None none.ConfigClient
}

type Layers struct {
  processor processor.Processor
  channel   channel.Channel

  // Chans for handling closing of the covert channel
  readClose chan interface{}
  readCloseDone chan interface{}
}

type Controller struct {
  config configData
  layers *Layers
  upgrader websocket.Upgrader
  clients map[*websocket.Conn]bool
  clientLock sync.Mutex
  waitGroup sync.WaitGroup
  clientStop chan interface{}
  recvStop   chan interface{}
  sendStop   chan interface{}
  doneWsSend chan interface{}
  doneWsRecv chan interface{}
  wsSend chan []byte
  wsRecv chan []byte
}
