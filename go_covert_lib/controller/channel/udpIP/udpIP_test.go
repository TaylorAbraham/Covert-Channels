package udpIP

import (
	"bytes"
	"log"
	"testing"
	"time"
)

var sconfTimeout Config = Config{
	FriendIP:          [4]byte{127, 0, 0, 1},
	OriginIP:          [4]byte{127, 0, 0, 1},
	FriendReceivePort: 8080,
	OriginReceivePort: 8081,
	DialTimeout:       time.Second,
	AcceptTimeout:     time.Second,
	ReadTimeout:       time.Second,
	WriteTimeout:      time.Second,
}

var rconfTimeout Config = Config{
	FriendIP:          [4]byte{127, 0, 0, 1},
	OriginIP:          [4]byte{127, 0, 0, 1},
	FriendReceivePort: 8081,
	OriginReceivePort: 8080,
	DialTimeout:       time.Second,
	AcceptTimeout:     time.Second,
	ReadTimeout:       time.Second,
	WriteTimeout:      time.Second,
}

func TestReceiveSend(t *testing.T) {

	log.Println("Starting TestReceiveSend")

	sch, err := MakeChannel(sconfTimeout)
	if err != nil {
		t.Errorf("err = '%s'; want nil", err.Error())
	}

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

	for _, input := range inputs {
		go func() {
			var data [15]byte
			nr, rErr = rch.Receive(data[:])
			select {
			case c <- data[:nr]:
			case <-time.After(time.Second * 5):
			}
		}()

		sendAndCheck(t, input, sch)

		receiveAndCheck(t, input, c)

		if rErr != nil {
			t.Errorf("err = '%s'; want nil", rErr.Error())
		}
	}
	if err := sch.Close(); err != nil {
		t.Errorf("err = '%s'; want nil", err.Error())
	}
	if err := rch.Close(); err != nil {
		t.Errorf("err = '%s'; want nil", err.Error())
	}
}

func TestReceiveSendSelf(t *testing.T) {

	log.Println("Starting TestReceiveSendSelf")

	var conf Config = Config{
		FriendIP:          [4]byte{127, 0, 0, 1},
		OriginIP:          [4]byte{127, 0, 0, 1},
		FriendReceivePort: 8080,
		OriginReceivePort: 8080,
	}

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

	for _, input := range inputs {
		go func() {
			var data [15]byte
			nr, rErr = ch.Receive(data[:])
			select {
			case c <- data[:nr]:
			case <-time.After(time.Second * 5):
			}
		}()

		sendAndCheck(t, input, ch)

		receiveAndCheck(t, input, c)

		if rErr != nil {
			t.Errorf("err = '%s'; want nil", rErr.Error())
		}
	}
	if err := ch.Close(); err != nil {
		t.Errorf("err = '%s'; want nil", err.Error())
	}

}

func sendAndCheck(t *testing.T, input []byte, sch *Channel) {
	n, err := sch.Send(input)
	if err != nil {
		t.Errorf("err = '%s'; want nil", err.Error())
	}
	if n != uint64(len(input)) {
		t.Errorf("send n = %d; want %d", n, len(input))
	}
}

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
