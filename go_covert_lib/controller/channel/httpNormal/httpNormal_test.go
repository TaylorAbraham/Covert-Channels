package httpNormal

import (
	"bytes"
	"log"
	"testing"
	"time"
)

var sconf Config = Config{
	FriendIP:       [4]byte{127, 0, 0, 1},
	OriginIP:       [4]byte{127, 0, 0, 1},
	FriendPort:     2001,
	OriginPort:     2000,
	UserType:       Server,
	ClientPollRate: time.Millisecond * 500,
	ClientTimeout:  2 * time.Second,
}

var rconf Config = Config{
	FriendIP:       [4]byte{127, 0, 0, 1},
	OriginIP:       [4]byte{127, 0, 0, 1},
	FriendPort:     2000,
	OriginPort:     2001,
	UserType:       Client,
	ClientPollRate: time.Millisecond * 500,
	ClientTimeout:  2 * time.Second,
}

func TestReceiveSendGet(t *testing.T) {

	log.Println("Starting TestReceiveSend")

	sch, err := MakeChannel(sconf)
	if err != nil {
		t.Errorf("err = '%s'; want nil", err.Error())
	}

	rch, err := MakeChannel(rconf)
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

func TestReceiveSendPost(t *testing.T) {

	log.Println("Starting TestReceiveSend")

	sconfcopy := sconf
	sconfcopy.FriendPort = 3001
	sconfcopy.OriginPort = 3002
	sconfcopy.UserType = Server

	sch, err := MakeChannel(sconfcopy)
	if err != nil {
		t.Errorf("err = '%s'; want nil", err.Error())
	}

	rconfcopy := rconf
	rconfcopy.FriendPort = 3002
	rconfcopy.OriginPort = 3001
	rconfcopy.UserType = Client

	rch, err := MakeChannel(rconfcopy)
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
	case <-time.After(time.Second * 5):
		t.Errorf("Read timeout")
	}
}
