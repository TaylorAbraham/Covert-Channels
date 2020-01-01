package tcp

import (
	"bytes"
	"math/rand"
	"testing"
	"time"
)

var sconf Config = Config{
	FriendIP:          [4]byte{127, 0, 0, 1},
	OriginIP:          [4]byte{127, 0, 0, 1},
	FriendReceivePort: 8080,
	OriginReceivePort: 8081,
}

var rconf Config = Config{
	FriendIP:          [4]byte{127, 0, 0, 1},
	OriginIP:          [4]byte{127, 0, 0, 1},
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
		c    chan []byte = make(chan []byte)
		rErr error
		nr   uint64
		// Test with message with many characters and with 0 characters
		inputs [][]byte = [][]byte{[]byte("Hello world!"), []byte("")}
	)

	for _, input := range inputs {
		go func() {
			var data [15]byte
			nr, rErr = rch.Receive(data[:], nil)
			c <- data[:nr]
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
		input []byte = []byte("123456")
		rErr  error
		nr    uint64
		c     chan []byte = make(chan []byte)
	)

	go func() {
		var data [5]byte
		nr, rErr = rch.Receive(data[:], nil)
		c <- data[:nr]
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

func randomString() string {
	r := rand.New(rand.NewSource(time.Now().UnixNano()))
	l := r.Int() & 0xFF
	buf := []byte{}
	for i := 0; i < l; i++ {
		buf = append(buf, byte(r.Int()&0xFF))
	}
	return string(buf)
}

func TestMultipleSend(t *testing.T) {
	runMultiTest(t, sconf, rconf)
}

func TestMultipleSendTimeout(t *testing.T) {
	sconfCopy := sconf
	sconfCopy.DialTimeout = time.Second
	sconfCopy.ReadTimeout = time.Second
	rconfCopy := rconf
	rconfCopy.DialTimeout = time.Second
	rconfCopy.ReadTimeout = time.Second
	runMultiTest(t, sconfCopy, rconfCopy)
}

func runMultiTest(t *testing.T, sconf, rconf Config) {

	sch, err := MakeChannel(sconf)
	if err != nil {
		t.Errorf("err = '%s'; want nil", err.Error())
	}

	rch, err := MakeChannel(rconf)
	if err != nil {
		t.Errorf("err = '%s'; want nil", err.Error())
	}

	var (
		rErr error
		nr   uint64
		// Test with message with many characters and with 0 characters
		inputs [][]byte = [][]byte{}
	)
	// Randomly generate input strings
	for i := 0; i < 32; i++ {
		inputs = append(inputs, []byte(randomString()))
	}

	for _, input := range inputs {
		sendAndCheck(t, input, sch)
	}

	for _, input := range inputs {
		var data [1024]byte
		nr, rErr = rch.Receive(data[:], nil)
		if rErr != nil {
			t.Errorf("err = '%s'; want nil", rErr.Error())
		} else if bytes.Compare(data[:nr], input) != 0 {
			t.Errorf("Received '%s'; want '%s'", string(data[:nr]), string(input))
			t.Error(data[:nr])
			t.Error(input)
		}
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
func sendAndCheck(t *testing.T, input []byte, sch *Channel) {
	n, err := sch.Send(input, nil)
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
