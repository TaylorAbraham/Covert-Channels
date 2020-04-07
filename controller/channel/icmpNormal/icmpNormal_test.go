package icmpNormal

import (
	"bytes"
	"log"
	"testing"
	"time"
)

var sconf Config = Config{
	FriendIP:        [4]byte{127, 0, 0, 1},
	OriginIP:        [4]byte{127, 0, 0, 1},
	DestinationPort: 8080,
	OriginPort:      8081,
}

var rconf Config = Config{
	FriendIP:        [4]byte{127, 0, 0, 1},
	OriginIP:        [4]byte{127, 0, 0, 1},
	DestinationPort: 8081,
	OriginPort:      8080,
}

// testing sending and receiving feature of the covert channel
func TestReceiveSend(t *testing.T) {

	log.Println("Starting ICMP TestReceiveSend")

	// create the ICMP channel
	sch, err := MakeChannel(sconf)
	if err != nil {
		t.Errorf("err = '%s'; want nil", err.Error())
	}

	rch, err := MakeChannel(rconf)
	if err != nil {
		t.Errorf("err = '%s'; want nil", err.Error())
	}

	var (
		c      chan []byte = make(chan []byte)
		rErr   error
		nr     uint64
		inputs [][]byte = [][]byte{[]byte("This is a normal channel"), []byte("")}
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

		// send the packet
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
