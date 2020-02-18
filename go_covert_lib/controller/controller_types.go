package controller

import (
	"sync"

	"./channel"
	"./channel/tcpHandshake"
	"./channel/tcpNormal"
	"./channel/tcpSyn"
	"./processor"
	"./processor/asymmetricEncryption"
	"./processor/caesar"
	"./processor/checksum"
	"./processor/gZipCompression"
	"./processor/none"
	"./processor/symmetricEncryption"
	"./processor/zLibCompression"

	"github.com/gorilla/websocket"
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
	OpCode  string
	Message string
}

type defaultConfig struct {
	Processor processorData
	Channel   channelData
}

type configData struct {
	OpCode     string
	Default    defaultConfig
	Processors []processorConfig
	Channel    channelConfig
}

type processorConfig struct {
	Type string
	Data processorData
}

type channelConfig struct {
	Type string
	Data channelData
}

type channelData struct {
	TcpSyn       tcpSyn.ConfigClient
	TcpHandshake tcpHandshake.ConfigClient
	TcpNormal    tcpNormal.ConfigClient
}

type processorData struct {
	None                 none.ConfigClient
	Caesar               caesar.ConfigClient
	Checksum             checksum.ConfigClient
	SymmetricEncryption  symmetricEncryption.ConfigClient
	AsymmetricEncryption asymmetricEncryption.ConfigClient
	GZipCompression      gZipCompression.ConfigClient
	ZLibCompression      zLibCompression.ConfigClient
}

type Layers struct {
	processors []processor.Processor
	channel    channel.Channel

	// Chans for handling closing of the covert channel
	readClose     chan interface{}
	readCloseDone chan interface{}
}

type Controller struct {
	config     configData
	layers     *Layers
	upgrader   websocket.Upgrader
	clients    map[*websocket.Conn]bool
	clientLock sync.Mutex
	waitGroup  sync.WaitGroup
	clientStop chan interface{}
	recvStop   chan interface{}
	sendStop   chan interface{}
	doneWsSend chan interface{}
	doneWsRecv chan interface{}
	wsSend     chan []byte
	wsRecv     chan []byte
}
