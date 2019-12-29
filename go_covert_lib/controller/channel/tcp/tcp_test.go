package tcp

import (
	"testing"
	"time"
)

var sconf Config = Config{
	FriendIP:   [4]byte{127, 0, 0, 1},
	OriginIP:   [4]byte{127, 0, 0, 1},
	FriendReceivePort: 8080,
	OriginReceivePort: 8081,
}

var rconf Config = Config{
	FriendIP:   [4]byte{127, 0, 0, 1},
	OriginIP:   [4]byte{127, 0, 0, 1},
	FriendReceivePort: 8081,
	OriginReceivePort: 8080,
}

func TestReceiveSend(t *testing.T) {

	sch, err := MakeChannel(sconf)
	if err != nil {
		t.Errorf("err = '%s'; want nil", err.Error())
	}

	rch, err := MakeChannel(rconf)
	if err != nil {
		t.Errorf("err = '%s'; want nil", err.Error())
	}

	var (
		c    chan string = make(chan string)
		rErr error
		nr   uint64
		// Test with message with many characters and with 0 characters
		inputs []string = []string{"Hello world!", ""}
	)

	for _, input := range inputs {
		go func() {
			var data [15]byte
			nr, rErr = rch.Receive(data[:], nil)
			c <- string(data[:nr])
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

func TestReceiveOverflow(t *testing.T) {

	sch, err := MakeChannel(sconf)
	if err != nil {
		t.Errorf("err = '%s'; want nil", err.Error())
	}

	rch, err := MakeChannel(rconf)
	if err != nil {
		t.Errorf("err = '%s'; want nil", err.Error())
	}

	var (
		input string = "123456"
		rErr  error
		nr    uint64
		c     chan string = make(chan string)
	)

	go func() {
		var data [5]byte
		nr, rErr = rch.Receive(data[:], nil)
		c <- string(data[:nr])
	}()

	sendAndCheck(t, input, sch)

	receiveAndCheck(t, input[:5], c)

	if rErr == nil {
		t.Errorf("err = nil; want error message")
	} else if rErr.Error() != "Buffer Full" {
		t.Errorf("err = '%s'; want 'Buffer Full'", rErr.Error())
	}

	if err := sch.Close(); err != nil {
		t.Errorf("err = '%s'; want nil", err.Error())
	}
	if err := rch.Close(); err != nil {
		t.Errorf("err = '%s'; want nil", err.Error())
	}
}
/*
func TestReceiveNone(t *testing.T) {

	rch, err := MakeChannel(rconf)
	if err != nil {
		t.Errorf("err = '%s'; want nil", err.Error())
	}

	var nr uint64

	var data [5]byte
	nr, err = rch.Receive(data[0:0], nil)
	if err != nil {
		t.Errorf("err = '%s'; want nil", err.Error())
	}
	if nr != 0 {
		t.Errorf("send n = %d; want %d", nr, 0)
	}

	if err := rch.Close(); err != nil {
		t.Errorf("err = '%s'; want nil", err.Error())
	}
}
*/
func sendAndCheck(t *testing.T, input string, sch *Channel) {
	n, err := sch.Send([]byte(input), nil)
	if err != nil {
		t.Errorf("err = '%s'; want nil", err.Error())
	}
	if n != uint64(len(input)) {
		t.Errorf("send n = %d; want %d", n, len(input))
	}
}

func receiveAndCheck(t *testing.T, input string, c chan string) {
	select {
	case received := <-c:
		if received != input {
			t.Errorf("received = %s; want %s", received, input)
		}
	case <-time.After(time.Millisecond * 500):
		t.Errorf("Read timeout")
	}
}
