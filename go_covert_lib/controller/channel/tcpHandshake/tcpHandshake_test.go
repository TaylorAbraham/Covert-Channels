package tcpHandshake

import (
	"bytes"
	"encoding/json"
	"io/ioutil"
	"log"
	"math/rand"
	"sort"
	"testing"
	"time"
)

// These tests must be run in sudo mode
// When running tests network traffic should be kept to a minimum (i.e. close an video streaming going on)
// The raw socket used to receive the covert packets has a default buffer that holds roughly 300 packets
// (based on my experimental results)
// If there is a lot of tcp traffic then packets involved in the test communication may be dropped,
// causing the message to not be received.
// This issue will manifest as a failure of either TestMultipleSend or TestMultipleSendTimeout.
// The received and transmitted byte string will be printed. The received byte will typically be empty
// (if the TCP handshake packets were dropped) or will match sections of the sent message (typically
// the start and end of the receive byte string will match the start and end of the sent bytes, with
// a region in the middle ommitted)
// If the test fails for reasons other than the case described above, then you should
// investigate further to identify any other potential bugs in the covert channel.
// Even with a video playing over the network, the above error is relatively uncommon,
// so an increase in frequence of the error should also be investigated.
// Keep in mind that the TestMultipleSend or TestMultipleSendTimeout tests are used to
// assess the performance when the covert channel is approaching (but not exceeding)
// its maximum storage capacity for number of connections. Thus they are not representative
// of the typical behaviour, where there should only be one connection at a time.

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

func TestReceiveOverflow(t *testing.T) {

	log.Println("Starting TestReceiveOverflow")

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
		nr, rErr = rch.Receive(data[:])
		select {
		case c <- data[:nr]:
		case <-time.After(time.Second * 5):
		}
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

func randomString(maxLen int) string {
	r := rand.New(rand.NewSource(time.Now().UnixNano()))
	l := r.Int() & maxLen
	buf := []byte{}
	for i := 0; i < l; i++ {
		buf = append(buf, byte(r.Int()&0xFF))
	}
	return string(buf)
}

func TestMultipleSend(t *testing.T) {
	log.Println("Starting TestMultipleSend")
	sconfCopy := sconf
	sconfCopy.logPackets = true
	rconfCopy := rconf
	rconfCopy.logPackets = true
	runMultiTest(t, sconf, rconf)
}

func TestMultipleSendTimeout(t *testing.T) {
	log.Println("Starting TestMultipleSendTimeout")
	sconfCopy := sconfTimeout
	sconfCopy.logPackets = true
	rconfCopy := rconfTimeout
	rconfCopy.logPackets = true
	runMultiTest(t, sconfCopy, rconfCopy)
}

// Test that multiple messages can be sent before a
// the receive calls its Receive method
func runMultiTest(t *testing.T, sconf, rconf Config) {

	sch, err := MakeChannel(sconf)
	if err != nil {
		t.Errorf("err = '%s'; want nil", err.Error())
	}

	rch, err := MakeChannel(rconf)
	if err != nil {
		t.Errorf("err = '%s'; want nil", err.Error())
	}

	type opOutput struct {
		size uint64
		sent uint64
		data []byte
		err  error
	}

	var (
		// Test with message with many characters and with 0 characters
		inputs   []string
		received chan opOutput = make(chan opOutput)
		sended   chan opOutput = make(chan opOutput)
	)

	// Randomly generate input strings
	for i := 0; i < 32; i++ {
		inputs = append(inputs, randomString(0xFF))
		// Test simultaneous receives
		go func() {
			var data [1024]byte
			var rOut opOutput
			nr, err := rch.Receive(data[:])
			rOut.err = err
			rOut.data = data[:nr]
			select {
			case received <- rOut:
			case <-time.After(time.Second * 5):
			}
		}()
	}

	for _, input := range inputs {
		// Test simultaneous sends
		go func(input []byte) {
			var sOut opOutput
			nr, err := sch.Send(input)
			sOut.err = err
			sOut.sent = nr
			sOut.size = uint64(len(input))
			select {
			case sended <- sOut:
			case <-time.After(time.Second * 5):
			}
		}([]byte(input))
	}

	var outputs []string

	for i := 0; i < 32; i++ {
		select {
		case sOut := <-sended:
			if sOut.err != nil {
				t.Errorf("err = '%s'; want nil", sOut.err)
			}
			if sOut.sent != sOut.size {
				t.Errorf("send n = %d; want %d", sOut.sent, sOut.size)
			}
		case <-time.After(time.Second * 5):
			t.Errorf("Receive Timeout")
		}

		select {
		case rOut := <-received:
			if rOut.err != nil {
				t.Errorf("err = '%s'; want nil", rOut.err)
			}
			outputs = append(outputs, string(rOut.data))
		case <-time.After(time.Second * 5):
			t.Errorf("Receive Timeout")
		}
	}

	sort.Strings(outputs)
	sort.Strings(inputs)

	logPkts := func(mp map[uint16][]packet, name string) {
		if buf, err := json.Marshal(mp); err == nil {
			if err := ioutil.WriteFile(name, buf, 0777); err != nil {
				log.Println("Could not write file")
			}
		} else {
			log.Println("Could not marshal")
		}
	}

	if len(inputs) == len(outputs) {
		for i := range inputs {
			if outputs[i] != inputs[i] {
				t.Errorf("Received '%s'; want '%s'", outputs[i], inputs[i])
				t.Error([]byte(outputs[i]))
				t.Error([]byte(inputs[i]))
				logPkts(sch.sendPktLog.pktMap, "./sendLogMismatch")
				logPkts(rch.receivePktLog.pktMap, "./receiveLogMismatch")
				break
			}
		}
	} else {
		t.Errorf("Insufficent replies received")
		logPkts(sch.sendPktLog.pktMap, "./sendLogTimeout")
		logPkts(rch.receivePktLog.pktMap, "./receiveLogTimeout")
	}

	if err := sch.Close(); err != nil {
		t.Errorf("err = '%s'; want nil", err.Error())
	}
	if err := rch.Close(); err != nil {
		t.Errorf("err = '%s'; want nil", err.Error())
	}
}

// Test that no panic errors occur
// with large numbers of messages being sent
func TestStress(t *testing.T) {
	log.Println("Starting TestStress")

	sch, err := MakeChannel(sconfTimeout)
	if err != nil {
		t.Errorf("err = '%s'; want nil", err.Error())
	}

	rch, err := MakeChannel(rconfTimeout)
	if err != nil {
		t.Errorf("err = '%s'; want nil", err.Error())
	}

	var (
		// Test with message with many characters and with 0 characters
		inputs [][]byte = [][]byte{}
	)
	var done chan int = make(chan int)

	// Randomly generate input strings
	// Send more messages than can be stored simultaneously
	for i := 0; i < 64; i++ {
		var s string = randomString(0x2FF)
		// Send some strings that are longer than the message buffer size
		inputs = append(inputs, []byte(s))

		// simultaneous reads should be safe
		go func(i int) {
			// Choose a length between the buffer size of the go channels
			// that hold the incoming message and the largest possible random string
			// specified above
			var data [700]byte
			data[0] = byte(i)
			rch.Receive(data[:])
			done <- i
		}(i)
	}

	for _, input := range inputs {
		sch.Send(input)
	}

	if err := sch.Close(); err != nil {
		t.Errorf("err = '%s'; want nil", err.Error())
	}
	if err := rch.Close(); err != nil {
		t.Errorf("err = '%s'; want nil", err.Error())
	}

	var dones map[int]bool = make(map[int]bool)

	for i := 0; i < 64; i++ {
		select {
		case ri := <-done:
			dones[ri] = true
		case <-time.After(time.Second * 15):
		}
	}

	for i := 0; i < 64; i++ {
		if _, ok := dones[i]; !ok {
			t.Errorf("Goroutine %d did not complete", i)
		}
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

func TestClose(t *testing.T) {
	log.Println("Starting TestClose")

	sch, err := MakeChannel(sconf)
	if err != nil {
		t.Errorf("err = '%s'; want nil", err.Error())
	}

	rch, err := MakeChannel(rconf)
	if err != nil {
		t.Errorf("err = '%s'; want nil", err.Error())
	}

	buf := make([]byte, 32)
	doneClose := make(chan bool)

	go func() {
		time.Sleep(time.Millisecond * 100)
		rch.Close()
		sch.Close()
		doneClose <- true
	}()

	n, err := rch.Receive(buf)
	if n != 0 {
		t.Errorf("Expected 0 bytes, got %d", n)
	} else if err == nil {
		t.Errorf("Expected error; got nil")
	}

	select {
	case <-doneClose:
	case <-time.After(time.Second):
		t.Errorf("Close timeout")
	}

	n, err = sch.Send(buf)
	if n != 0 {
		t.Errorf("Expected 0 bytes, got %d", n)
	} else if err == nil {
		t.Errorf("Expected error; got nil")
	}

	n, err = rch.Receive(buf)
	if n != 0 {
		t.Errorf("Expected 0 bytes, got %d", n)
	} else if err == nil {
		t.Errorf("Expected error; got nil")
	}

}

// Test that closing can happen at random times
// and any waiting functions will still return with an error
func TestCloseMultiple(t *testing.T) {
	log.Println("Starting TestCloseMultiple")

	var sChans []*Channel = make([]*Channel, 5)
	var rChans []*Channel = make([]*Channel, 5)
	var closers []chan bool = make([]chan bool, 5)
	var sDones []chan bool = make([]chan bool, 5)
	var rDones []chan bool = make([]chan bool, 5)

	sconfCopy := sconf
	rconfCopy := rconf

	var err error

	for i := 0; i < 5; i++ {
		sconfCopy.OriginReceivePort += 10
		sconfCopy.FriendReceivePort += 10
		rconfCopy.OriginReceivePort += 10
		rconfCopy.FriendReceivePort += 10

		sChans[i], err = MakeChannel(sconfCopy)
		if err != nil {
			t.Errorf("err = '%s'; want nil", err.Error())
		}
		rChans[i], err = MakeChannel(rconfCopy)
		if err != nil {
			t.Errorf("err = '%s'; want nil", err.Error())
		}
		closers[i] = make(chan bool)
		sDones[i] = make(chan bool)
		rDones[i] = make(chan bool)

		go func(c *Channel, cl, dn chan bool) {
			buf := make([]byte, 20)
			for {
				_, err := c.Receive(buf)
				select {
				case <-cl:
					if err != nil {
						close(dn)
						return
					}
				default:
				}
			}
		}(rChans[i], closers[i], rDones[i])
		go func(c *Channel, cl, dn chan bool) {
			buf := make([]byte, 10)
			for {
				_, err := c.Send(buf)
				select {
				case <-cl:
					if err != nil {
						close(dn)
						return
					}
				default:
				}
			}
		}(sChans[i], closers[i], sDones[i])
	}

	for i := 0; i < 5; i++ {
		go func(cl chan bool, sc, rc *Channel) {
			r := rand.New(rand.NewSource(time.Now().UnixNano()))
			delay := uint32(r.Int()) % 300
			time.Sleep(time.Duration(delay) * time.Millisecond)
			close(cl)
			sc.Close()
			rc.Close()
		}(closers[i], sChans[i], rChans[i])
	}

	for i := 0; i < 5; i++ {
		select {
		case <-rDones[i]:
		case <- time.After(time.Second * 2):
			t.Errorf("Close Timeout")
		}
		select {
		case <-sDones[i]:
		case <- time.After(time.Second * 2):
			t.Errorf("Close Timeout")
		}
	}

}
