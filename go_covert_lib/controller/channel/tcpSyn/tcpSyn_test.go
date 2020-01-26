package tcpSyn

import (
	"testing"
	"time"
)

var sconf Config = Config{
	FriendIP:   [4]byte{127, 0, 0, 1},
	OriginIP:   [4]byte{127, 0, 0, 1},
	FriendPort: 8080,
	OriginPort: 8081,
	Delimiter:  Protocol,
}

var rconf Config = Config{
	FriendIP:   [4]byte{127, 0, 0, 1},
	OriginIP:   [4]byte{127, 0, 0, 1},
	FriendPort: 8081,
	OriginPort: 8080,
	Delimiter:  Protocol,
}

func TestProtocolDelimiter(t *testing.T) {

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
			nr, rErr = rch.Receive(data[:])
			c <- string(data[:nr])
		}()

		sendAndCheck(t, input, sch)

		receiveAndCheck(t, input, c)

		if rErr != nil {
			t.Errorf("err = '%s'; want nil", rErr.Error())
		}
	}
}

func TestProtocolDelimiterOverflow(t *testing.T) {

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
		nr, rErr = rch.Receive(data[:])
		c <- string(data[:nr])
	}()

	sendAndCheck(t, input, sch)

	receiveAndCheck(t, input[:5], c)

	if rErr == nil {
		t.Errorf("err = nil; want error message")
	} else if rErr.Error() != "End packet not received, buffer full" {
		t.Errorf("err = '%s'; want 'End packet not received, buffer full'", rErr.Error())
	}
}

func TestReceiveNone(t *testing.T) {

	rch, err := MakeChannel(rconf)
	if err != nil {
		t.Errorf("err = '%s'; want nil", err.Error())
	}

	var nr uint64

	var data [5]byte
	nr, err = rch.Receive(data[0:0])
	if err != nil {
		t.Errorf("err = '%s'; want nil", err.Error())
	}
	if nr != 0 {
		t.Errorf("send n = %d; want %d", nr, 0)
	}
}

func sendAndCheck(t *testing.T, input string, sch *Channel) {
	n, err := sch.Send([]byte(input))
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
