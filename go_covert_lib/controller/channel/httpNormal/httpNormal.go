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
const (
	Client = 0
	Server = 1
)

// This is a normal, non-covert HTTP messaging channel
// The message is sent using normal HTTP packets with the proper OS tcp functions
type Config struct {
	FriendIP   [4]byte
	OriginIP   [4]byte
	FriendPort uint16
	OriginPort uint16
	UserType   uint8

	ClientPollRate time.Duration
	ClientTimeout  time.Duration
	ReadTimeout    time.Duration
	WriteTimeout   time.Duration
}

// A HTTP covert channel
type Channel struct {
	conf Config
	srv  *http.Server
	clt  *http.Client

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

	// if the channel has been specified as the server
	if c.conf.UserType == Server {
		mux := http.NewServeMux()
		mux.HandleFunc("/", c.handleFunc)

		// launch a new http server
		c.srv = &http.Server{Handler: mux}

		// starts up the http server to listen
		l, err := net.Listen("tcp4", ":"+strconv.Itoa(int(c.conf.OriginPort)))
		if err != nil {
			return nil, err
		}

		go func() { c.srv.Serve(l) }()

		//buffers used to hold the server sent messages and received messages
		c.serverSendBuf = make(chan []byte, maxMsg)
		c.serverRecBuf = make(chan []byte, maxMsg)
	}
	return c, nil
}

//Handler function that is called everytime a new http request is received
func (c *Channel) handleFunc(w http.ResponseWriter, r *http.Request) {

	//if the request message is a get request
	//extract the data from the servers send buffer and send it to the client
	if r.Method == "GET" {
		select {
		case data := <-c.serverSendBuf:
			w.Header().Add("Valid", "0")
			io.WriteString(w, string(data))
		case <-time.After(time.Millisecond * 50):
			//log.Println("GET Request timeout")
			w.Header().Add("Valid", "1")
		case <-c.cancel:
			w.Header().Add("Valid", "1")
		}

		// if the request message is a post request
		// extract the data from the body of the request message and prepare it to be displayed to the user
	} else if r.Method == "POST" {
		buf := bytes.Buffer{}
		buf.ReadFrom(r.Body)
		select {
		case c.serverRecBuf <- buf.Bytes():
		case <-time.After(time.Millisecond * 50):
			//log.Println("POST Request timeout")
		case <-c.cancel:
		}
	} else {
		log.Println("Invalid Covert Message")
	}
}

// Close the covert the channel
func (c *Channel) Close() error {
	select {
	case <-c.cancel:
		return nil
	default:
		close(c.cancel)
		if c.conf.UserType == Server {
			ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
			defer cancel()
			return c.srv.Shutdown(ctx)
		}
	}
	return nil
}

//receive information from another computer in the form of data in a byte array
func (c *Channel) Receive(data []byte) (uint64, error) {

	// if the user is a server type computer
	if c.conf.UserType == Server {

		//Determines whether the channel is handling timeout or not
		//then read from the servers receive buffer
		//this is without timeout
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
			// this is with timeout
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

		// if the user is client type computer
	} else {
		n, success, err := c.clientRequest(data)
		if success || err != nil {
			return n, err
		} else {
			if c.conf.ClientPollRate > 0 {
				// Loop while making requests
				var ticker *time.Ticker = time.NewTicker(c.conf.ClientPollRate)
				defer ticker.Stop()
				if c.conf.ClientTimeout == time.Duration(0) {
					// No timeout
					for {
						select {
						case <-ticker.C:
							n, success, err := c.clientRequest(data)
							if success || err != nil {
								return n, err
							}
						case <-c.cancel:
							return 0, errors.New("Channel closed")
						}
					}
				} else {
					for {
						select {
						case <-ticker.C:
							n, success, err := c.clientRequest(data)
							if success || err != nil {
								return n, err
							}
						case <-time.After(c.conf.ClientTimeout):
							return 0, errors.New("Client Timeout")
						case <-c.cancel:
							return 0, errors.New("Channel closed")
						}
					}
				}
			} else {
				return n, err
			}
		}
	}
}

// send information to another computer in the form of data in a byte array
func (c *Channel) Send(data []byte) (uint64, error) {

	//copy the data into a new slice so that even if the original slice is modified,
	// the data sent on the channel will stay the same
	var dataCopy []byte = make([]byte, len(data))
	copy(dataCopy, data)

	// if the user is a server type computer
	if c.conf.UserType == Server {

		//Determines whether the channel is handling timeout or not
		//then read from the servers send buffer
		//this is without timeout
		if c.conf.WriteTimeout == 0 {
			select {
			case c.serverSendBuf <- dataCopy:
				return uint64(len(dataCopy)), nil
			case <-c.cancel:
				return 0, errors.New("Channel closed")
			}

			//this is with timeout
		} else {
			select {
			case c.serverSendBuf <- dataCopy:
				return uint64(len(dataCopy)), nil
			case <-time.After(c.conf.WriteTimeout):
				return 0, errors.New("Write Timeout")
			case <-c.cancel:
				return 0, errors.New("Channel closed")
			}
		}

		// if the user is client type computer
	} else {

		//post the http request message
		addr := &net.TCPAddr{IP: c.conf.FriendIP[:], Port: int(c.conf.FriendPort)}
		_, err := http.Post("http://"+addr.String()+"/", "text/plain", bytes.NewBuffer(dataCopy))

		//as long as there is no error
		//return an integer with the length of the data to be sent
		if err == nil {
			return uint64(len(dataCopy)), nil
		} else {
			return 0, err
		}
	}
}

func (c *Channel) clientRequest(data []byte) (uint64, bool, error) {
	addr := &net.TCPAddr{IP: c.conf.FriendIP[:], Port: int(c.conf.FriendPort)}
	resp, err := http.Get("http://" + addr.String() + "/")

	//as long as there is no error
	//extract the information from the body of the reponse message
	if err == nil {
		if resp.Header.Get("Valid") == "0" {
			buf := bytes.Buffer{}
			buf.ReadFrom(resp.Body)
			copy(data, buf.Bytes())
			if len(buf.Bytes()) > len(data) {
				return uint64(len(data)), true, errors.New("Buffer overflow")
			} else {
				return uint64(len(buf.Bytes())), true, nil
			}
		} else {
			return 0, false, nil
		}
	} else {
		return 0, false, err
	}
}
