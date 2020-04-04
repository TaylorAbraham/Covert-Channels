package tcpSyn

import (
	"../embedders"
	"bytes"
	"log"
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

var confList []Config = []Config{
	Config{
		FriendIP:   [4]byte{127, 0, 0, 1},
		OriginIP:   [4]byte{127, 0, 0, 1},
		FriendPort: 8080,
		OriginPort: 8081,
		Delimiter:  Protocol,
		Embedder:   &embedders.TcpIpSeqEncoder{},
	},
	Config{
		FriendIP:   [4]byte{127, 0, 0, 1},
		OriginIP:   [4]byte{127, 0, 0, 1},
		FriendPort: 8080,
		OriginPort: 8081,
		Delimiter:  Protocol,
		Embedder:   &embedders.TcpIpIDEncoder{},
	},
	Config{
		FriendIP:   [4]byte{127, 0, 0, 1},
		OriginIP:   [4]byte{127, 0, 0, 1},
		FriendPort: 8080,
		OriginPort: 8081,
		Delimiter:  Protocol,
		Embedder:   &embedders.TcpIpUrgPtrEncoder{},
	},
	Config{
		FriendIP:   [4]byte{127, 0, 0, 1},
		OriginIP:   [4]byte{127, 0, 0, 1},
		FriendPort: 8080,
		OriginPort: 8081,
		Delimiter:  Protocol,
		Embedder:   &embedders.TcpIpUrgFlgEncoder{},
	},
	Config{
		FriendIP:   [4]byte{127, 0, 0, 1},
		OriginIP:   [4]byte{127, 0, 0, 1},
		FriendPort: 8080,
		OriginPort: 8081,
		Delimiter:  Protocol,
		Embedder:   &embedders.TcpIpTimestampEncoder{},
	},
	Config{
		FriendIP:   [4]byte{127, 0, 0, 1},
		OriginIP:   [4]byte{127, 0, 0, 1},
		FriendPort: 8080,
		OriginPort: 8081,
		Delimiter:  Protocol,
		Embedder:   &embedders.TcpIpEcnEncoder{},
	},
	Config{
		FriendIP:   [4]byte{127, 0, 0, 1},
		OriginIP:   [4]byte{127, 0, 0, 1},
		FriendPort: 8080,
		OriginPort: 8081,
		Delimiter:  Protocol,
		Embedder:   &embedders.TcpIpTemporalEncoder{Emb: embedders.TemporalEncoder{time.Duration(50 * time.Millisecond)}},
	},
	Config{
		FriendIP:   [4]byte{127, 0, 0, 1},
		OriginIP:   [4]byte{127, 0, 0, 1},
		FriendPort: 8080,
		OriginPort: 8081,
		Delimiter:  Protocol,
		Embedder:   &embedders.TcpIpEcnTempEncoder{TmpEmb: embedders.TemporalEncoder{time.Duration(50 * time.Millisecond)}},
	},
	Config{
		FriendIP:   [4]byte{127, 0, 0, 1},
		OriginIP:   [4]byte{127, 0, 0, 1},
		FriendPort: 8080,
		OriginPort: 8081,
		Delimiter:  Protocol,
		Embedder:   &embedders.TcpIpFreqEncoder{},
	},
}

func TestSendReceive(t *testing.T) {

	log.Println("Starting TestReceiveSend")
	for i := range confList {
		confcp := confList[i]
		confcp.FriendPort = confList[i].OriginPort
		confcp.OriginPort = confList[i].FriendPort
		runSendReceiveTest(t, confList[i], confcp)
	}
}

func runSendReceiveTest(t *testing.T, conf1, conf2 Config) {

	sch, err := MakeChannel(conf1)
	if err != nil {
		t.Errorf("err = '%s'; want nil", err.Error())
		return
	}

	rch, err := MakeChannel(conf2)
	if err != nil {
		t.Errorf("err = '%s'; want nil", err.Error())
		return
	}

	messageExcange(t, sch, rch, [][]byte{[]byte("Hello world!"), []byte(""), []byte("A"), []byte("Hello\nworld!"), []byte("🍌🍌🍌")})
}

func messageExcange(t *testing.T, ch1, ch2 *Channel, inputs [][]byte) {
	for _, input := range inputs {
		sendReceive(t, ch1, ch2, input)
		sendReceive(t, ch2, ch1, input)
	}
	if err := ch1.Close(); err != nil {
		t.Errorf("err = '%s'; want nil", err.Error())
	}
	if err := ch2.Close(); err != nil {
		t.Errorf("err = '%s'; want nil", err.Error())
	}
}

func sendReceive(t *testing.T, sch, rch *Channel, input []byte) {
	var (
		c    chan []byte = make(chan []byte)
		rErr error
		nr   uint64
	)
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

func TestProtocolDelimiterOverflow(t *testing.T) {

	log.Println("Starting Protocol Delimiter Overflow")
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
		c     chan []byte = make(chan []byte)
	)

	go func() {
		var data [5]byte
		nr, rErr = rch.Receive(data[:])
		c <- data[:nr]
	}()

	sendAndCheck(t, []byte(input), sch)

	receiveAndCheck(t, []byte(input)[:5], c)

	if rErr == nil {
		t.Errorf("err = nil; want error message")
	} else if rErr.Error() != "End packet not received, buffer full" {
		t.Errorf("err = '%s'; want 'End packet not received, buffer full'", rErr.Error())
	}
}

func TestReceiveNone(t *testing.T) {

	log.Println("Starting Buffer Delimiter Receive None")
	rconfCopy := rconf
	rconfCopy.Delimiter = Buffer
	rch, err := MakeChannel(rconfCopy)
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
