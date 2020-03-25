package icmpIP

import (
	"bytes"
	"log"
	"testing"
	"time"
)

var sconfTimeout Config = Config{
	FriendIP:     [4]byte{127, 0, 0, 1},
	OriginIP:     [4]byte{127, 0, 0, 1},
	Identifier:   1234,
	ReadTimeout:  time.Second,
	WriteTimeout: time.Second,
}

var rconfTimeout Config = Config{
	FriendIP:     [4]byte{127, 0, 0, 1},
	OriginIP:     [4]byte{127, 0, 0, 1},
	Identifier:   1234,
	ReadTimeout:  time.Second,
	WriteTimeout: time.Second,
}

// testing sending and receiving feature of the covert channel
func TestReceiveSend(t *testing.T) {

	log.Println("Starting ICMP covert TestReceiveSend")

	sch, err := MakeChannel(sconfTimeout)
	if err != nil {
		t.Errorf("err = '%s'; want nil", err.Error())
	}

	// create the ICMP covert channel
	rch, err := MakeChannel(rconfTimeout)
	if err != nil {
		t.Errorf("err = '%s'; want nil", err.Error())
	}

	var (
		c    chan []byte = make(chan []byte)
		rErr error
		nr   uint64
		// Test with message with many characters and with 0 characters
		inputs [][]byte = [][]byte{[]byte("Hello world!"), []byte("")}
	)

	// construct the packets
	for _, input := range inputs {
		go func() {
			var data [50]byte
			nr, rErr = rch.Receive(data[:])
			select {
			case c <- data[:nr]:
			case <-time.After(time.Second * 5):
			}
		}()

		// send the packets
		sendAndCheck(t, input, sch)

		// receive the packet
		receiveAndCheck(t, input, c)

		if rErr != nil {
			t.Errorf("err = '%s'; want nil", rErr.Error())
		}
	}

	// close the channel
	if err := sch.Close(); err != nil {
		t.Errorf("err = '%s'; want nil", err.Error())
	}

	if err := rch.Close(); err != nil {
		t.Errorf("err = '%s'; want nil", err.Error())
	}
}

// test sending and receiving to the same IP
func TestReceiveSendSelf(t *testing.T) {

	log.Println("Starting ICMP covert TestReceiveSendSelf")

	// configure the addresses
	var conf Config = Config{
		FriendIP:   [4]byte{127, 0, 0, 1},
		OriginIP:   [4]byte{127, 0, 0, 1},
		Identifier: 8080,
	}

	// create the ICMP covert channel
	ch, err := MakeChannel(conf)
	if err != nil {
		t.Errorf("err = '%s'; want nil", err.Error())
	}

	var (
		c    chan []byte = make(chan []byte)
		rErr error
		nr   uint64
		// Test with message with many characters and with 0 characters
		inputs [][]byte = [][]byte{[]byte("Hello world!"), []byte("")}
	)

	// construct the packet
	for _, input := range inputs {
		go func() {
			var data [50]byte
			nr, rErr = ch.Receive(data[:])
			select {
			case c <- data[:nr]:
			case <-time.After(time.Second * 5):
			}
		}()

		// send the packet
		sendAndCheck(t, input, ch)

		// receive the packet
		receiveAndCheck(t, input, c)

		if rErr != nil {
			t.Errorf("err = '%s'; want nil", rErr.Error())
		}
	}

	// close the channel
	if err := ch.Close(); err != nil {
		t.Errorf("err = '%s'; want nil", err.Error())
	}
}

// just send the ICMP packets
func sendAndCheck(t *testing.T, input []byte, sch *Channel) {
	n, err := sch.Send(input)
	if err != nil {
		t.Errorf("err = '%s'; want nil", err.Error())
	}
	if n != uint64(len(input)) {
		t.Errorf("send n = %d; want %d", n, len(input))
	}
}

// just receive the IMCP packets
func receiveAndCheck(t *testing.T, input []byte, c chan []byte) {
	select {
	case received := <-c:
		if bytes.Compare(received, input) != 0 {
			t.Errorf("received = %s; want %s", string(received), string(input))
		}
	case <-time.After(time.Millisecond * 500):
		t.Errorf("Read timeout")
	}
}
