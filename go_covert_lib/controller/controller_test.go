package controller

import (
	"context"
	"encoding/json"
	"github.com/gorilla/websocket"
	"net/http"
	"reflect"
	"testing"
	"time"
)

func TestShutdown(t *testing.T) {
	ctr, err := CreateController()
	if err != nil {
		t.Errorf("Unexpected open error: %s", err.Error())
	}
	err = ctr.Shutdown()
	if err != nil {
		t.Errorf("Unexpected close error: %s", err.Error())
	}
}

func openConn(addr string, port string, ctr *Controller, t *testing.T) (chan []byte, chan []byte, chan interface{}, chan interface{}) {

	var (
		write chan []byte      = make(chan []byte, 32)
		read  chan []byte      = make(chan []byte, 32)
		stop  chan interface{} = make(chan interface{})
		done  chan interface{} = make(chan interface{})
	)

	mux := http.NewServeMux()
	mux.HandleFunc("/covert", ctr.HandleFunc)

	srv := &http.Server{Addr: ":" + port, Handler: mux}

	go func() {
		srv.ListenAndServe()
		close(done)
	}()
	// Give time for the server to start
	time.Sleep(time.Millisecond * 100)
	client, _, err := websocket.DefaultDialer.Dial(addr, nil)
	if err != nil {
		t.Errorf("Unexpected dial error: %s", err.Error())
		return write, read, stop, done
	}

	go func() {
	loop:
		for {
			select {
			case data := <-write:
				client.WriteMessage(websocket.TextMessage, data)
			case <-stop:
				break loop
			}
		}
		if err = client.WriteMessage(websocket.CloseMessage, websocket.FormatCloseMessage(websocket.CloseNormalClosure, "")); err != nil {
			t.Errorf("Unexpected client close error: %s", err.Error())
		}

		if err = client.Close(); err != nil {
			t.Errorf("Unexpected client close error: %s", err.Error())
		}

		if err = ctr.Shutdown(); err != nil {
			t.Errorf("Unexpected controller close error: %s", err.Error())
		}
		ctx, _ := context.WithTimeout(context.Background(), 5*time.Second)
		srv.Shutdown(ctx)
	}()

	go func() {
	loop:
		for {
			select {
			case <-stop:
				break loop
			default:
			}
			_, data, err := client.ReadMessage()
			if err != nil {
				break
			}
			read <- data
		}
	}()

	return write, read, stop, done
}

func TestRetrieveConfig(t *testing.T) {
	ctr, _ := CreateController()

	write, read, stop, done := openConn("ws://127.0.0.1:9020/covert", "9020", ctr, t)

	write <- []byte("{\"OpCode\" : \"config\"}")

	checkConfig(read, DefaultConfig(), t)

	checkClose(stop, done, t)
}

func TestRetrieveWiteMessage(t *testing.T) {
	ctr1, _ := CreateController()
	ctr2, _ := CreateController()

	write1, read1, stop1, done1 := openConn("ws://127.0.0.1:9030/covert", "9030", ctr1, t)
	write2, read2, stop2, done2 := openConn("ws://127.0.0.1:9040/covert", "9040", ctr2, t)

	write1 <- []byte("{\"OpCode\" : \"config\"}")

	conf := checkConfig(read1, DefaultConfig(), t)

	conf.OpCode = "open"
	conf.Channel.Ipv4TCP.FriendPort.Value = 8090
	conf.Channel.Ipv4TCP.OriginPort.Value = 8091
	writeTestMsg(write1, conf, t)

	conf.Channel.Ipv4TCP.FriendPort.Value = 8091
	conf.Channel.Ipv4TCP.OriginPort.Value = 8090
	writeTestMsg(write2, conf, t)

	checkMsgType(read1, "open", "Open success", t)
	checkMsgType(read2, "open", "Open success", t)

	write1 <- []byte("{\"OpCode\" : \"write\", \"Message\" : \"Hello World!\"}")
	checkMsgType(read2, "read", "Hello World!", t)

	checkClose(stop1, done1, t)
	checkClose(stop2, done2, t)
}

func checkClose(stop chan interface{}, done chan interface{}, t *testing.T) {
	close(stop)
	select {
	case <-done:
	case <-time.After(time.Second * 5):
		t.Errorf("Unexpected shutdown timeout")
	}
}

func writeTestMsg(ch chan []byte, d interface{}, t *testing.T) {
	if data, err := json.Marshal(d); err != nil {
		t.Errorf("Unexpected marshal error: %s", err.Error())
	} else {
		ch <- data
	}
}

func checkMsgType(ch chan []byte, opcode string, msg string, t *testing.T) {
	select {
	case data := <-ch:
		var mt messageType
		if err := json.Unmarshal(data, &mt); err != nil {
			t.Errorf("Unexpected unmarshal error: %s", err.Error())
		} else {
			if mt.OpCode != opcode {
				t.Errorf("Open message does not have correct opcode: %s, want %s", mt.OpCode, opcode)
			}
			if mt.Message != msg {
				t.Errorf("Open message does not have correct message: %s, want %s", mt.Message, msg)
			}
		}
	case <-time.After(time.Second * 5):
		t.Errorf("Unexpected read timeout")
	}
}

func checkConfig(ch chan []byte, expt configData, t *testing.T) configData {
	var conf configData
	select {
	case data := <-ch:
		if err := json.Unmarshal(data, &conf); err != nil {
			t.Errorf("Unexpected unmarshal error: %s", err.Error())
		} else {
			if !reflect.DeepEqual(conf, expt) {
				t.Errorf("Configs do not match error: \n%v, \n%v", conf, expt)
			}
		}
	case <-time.After(time.Second * 5):
		t.Errorf("Unexpected read timeout")
	}
	return conf
}
