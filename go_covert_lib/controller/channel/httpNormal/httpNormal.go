package httpNormal

import (
	"bytes"
	"context"
	"errors"
	"io"
	"log"
	"net"
	"net/http"
	"strconv"
	"time"
)

const maxMsg = 32

// This is a normal, non-covert tcp messaging channel
// The message is sent using normal TCP packets with the proper OS tcp functions
type Config struct {
	FriendIP   [4]byte
	OriginIP   [4]byte
	FriendPort uint16
	OriginPort uint16
	UserType   bool

	ReadTimeout  time.Duration
	WriteTimeout time.Duration
}

// A TCP covert channel
type Channel struct {
	conf Config
	srv  *http.Server

	cancel chan bool

	serverSendBuf chan []byte
	serverRecBuf  chan []byte
}

// Create the covert channel, filling in the SeqEncoder
// with a default if none is provided
// Although the error is not yet used, it is anticipated
// that this function may one day be used for validating
// the data structure
func MakeChannel(conf Config) (*Channel, error) {
	c := &Channel{conf: conf, cancel: make(chan bool)}

	if conf.UserType {
		mux := http.NewServeMux()
		mux.HandleFunc("/", c.handleFunc)

		c.srv = &http.Server{ Handler: mux}

		l, err := net.Listen("tcp4", ":"+strconv.Itoa(int(c.conf.OriginPort)))
		if err != nil {
			return nil, err
		}

		go func() { c.srv.Serve(l) }()

		c.serverSendBuf = make(chan []byte, maxMsg)
		c.serverRecBuf = make(chan []byte, maxMsg)
	}
	return c, nil
}

func (c *Channel) handleFunc(w http.ResponseWriter, r *http.Request) {
	if r.Method == "GET" {
		select {
		case data := <-c.serverSendBuf:
			io.WriteString(w, string(data))
		case <-time.After(time.Millisecond * 50):
			log.Println("GET Request timeout")
		case <-c.cancel:
		}
	} else if r.Method == "POST" {
		buf := bytes.Buffer{}
		buf.ReadFrom(r.Body)
		select {
		case c.serverRecBuf <- buf.Bytes():
		case <-time.After(time.Millisecond * 50):
			log.Println("GET Request timeout")
		case <-c.cancel:
		}
	} else {
		log.Println("Invalid Covert Message")
	}
}

func (c *Channel) Close() error {
	select {
	case <-c.cancel:
		return nil
	default:
		close(c.cancel)
		if c.conf.UserType {
			ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
			defer cancel()
			return c.srv.Shutdown(ctx)
		}
	}
	return nil
}

func (c *Channel) Receive(data []byte) (uint64, error) {
	if c.conf.UserType {
		if c.conf.ReadTimeout == 0 {
			select {
			case httpdata := <-c.serverRecBuf:
				copy(data, httpdata)
				if len(httpdata) > len(data) {
					return uint64(len(data)), errors.New("Buffer overflow")
				} else {
					return uint64(len(httpdata)), nil
				}
			case <-c.cancel:
				return 0, errors.New("Channel closed")
			}
		} else {
			select {
			case httpdata := <-c.serverRecBuf:
				copy(data, httpdata)
				if len(httpdata) > len(data) {
					return uint64(len(data)), errors.New("Buffer overflow")
				} else {
					return uint64(len(httpdata)), nil
				}
			case <-time.After(c.conf.WriteTimeout):
				return 0, errors.New("Write Timeout")
			case <-c.cancel:
				return 0, errors.New("Channel closed")
			}
		}
	} else {
		addr := &net.TCPAddr{IP: c.conf.FriendIP[:], Port: int(c.conf.FriendPort)}
		resp, err := http.Get("http://" + addr.String() + "/")
		if err == nil {
			buf := bytes.Buffer{}
			buf.ReadFrom(resp.Body)
			copy(data, buf.Bytes())
			if len(buf.Bytes()) > len(data) {
				return uint64(len(data)), errors.New("Buffer overflow")
			} else {
				return uint64(len(buf.Bytes())), nil
			}
		} else {
			return 0, err
		}
	}
}

func (c *Channel) Send(data []byte) (uint64, error) {
	if c.conf.UserType {
		if c.conf.WriteTimeout == 0 {
			select {
			case c.serverSendBuf <- data:
				return uint64(len(data)), nil
			case <-c.cancel:
				return 0, errors.New("Channel closed")
			}
		} else {
			select {
			case c.serverSendBuf <- data:
				return uint64(len(data)), nil
			case <-time.After(c.conf.WriteTimeout):
				return 0, errors.New("Write Timeout")
			case <-c.cancel:
				return 0, errors.New("Channel closed")
			}
		}
	} else {
		addr := &net.TCPAddr{IP: c.conf.FriendIP[:], Port: int(c.conf.FriendPort)}
		_, err := http.Post("http://"+addr.String()+"/", "text/plain", bytes.NewBuffer(data))
		if err == nil {
			return uint64(len(data)), nil
		} else {
			return 0, err
		}
	}
}
