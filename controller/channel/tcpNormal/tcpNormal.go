package tcpNormal

import (
	"bytes"
	"context"
	"errors"
	"io"
	"log"
	"net"
	"os"
	"strconv"
	"time"
)

const maxAccept = 32

type acceptedConn struct {
	conn net.Conn
	// The TCP port used by our Friend IP in this covert message
	friendPort uint16
}

// This is a normal, non-covert tcp messaging channel
// The message is sent using normal TCP packets with the proper OS tcp functions
type Config struct {
	FriendIP          [4]byte
	OriginIP          [4]byte
	FriendReceivePort uint16
	OriginReceivePort uint16
	// The TCP dial timeout for the three way handshake in the the Send method
	DialTimeout time.Duration
	// The TCP accept timeout for the three way handshake in the the Receive method
	AcceptTimeout time.Duration
	// The intra-packet read timeout. Set zero for no timeout.
	// The receive method will block until a three way handshake
	// is complete and the listener returns, and will only exit with a
	// read timeout if the intra-packet delay is too large.
	ReadTimeout time.Duration
	// The timeout for writing the packet to a raw socket. Set zero for no timeout.
	WriteTimeout time.Duration
}

// A TCP covert channel
type Channel struct {
	conf     Config
	listener net.Listener

	cancel     chan bool
	acceptChan chan acceptedConn
}

// Create the covert channel, filling in the SeqEncoder
// with a default if none is provided
// Although the error is not yet used, it is anticipated
// that this function may one day be used for validating
// the data structure
func MakeChannel(conf Config) (*Channel, error) {
	var err error

	c := &Channel{conf: conf,
		cancel: make(chan bool),

		// Only 32 connections can be accepted before they begin to be dropped
		acceptChan: make(chan acceptedConn, maxAccept),
	}

	c.listener, err = net.Listen("tcp4", ":"+strconv.Itoa(int(c.conf.OriginReceivePort)))
	if err != nil {
		return nil, err
	}

	// A loop to accept incoming tcp connection
	go c.acceptLoop()

	return c, nil
}

func (c *Channel) Close() error {
	select {
	// Have we already closed
	case <-c.cancel:
		return nil
	default:
		close(c.cancel)
	}
	return c.listener.Close()
}

func (c *Channel) Receive(data []byte) (uint64, error) {

	var (
		ac acceptedConn
	)

	// Check if we should timeout
	if c.conf.AcceptTimeout == 0 {
		select {
		case ac = <-c.acceptChan:
		// Check if the covert channel has been closed
		case <-c.cancel:
			return 0, errors.New("Receive Cancelled")
		}
	} else {
		select {
		case ac = <-c.acceptChan:
		// Check if the covert channel has been closed
		case <-c.cancel:
			return 0, errors.New("Receive Cancelled")
		case <-time.After(c.conf.AcceptTimeout):
			return 0, errors.New("Accept timeout")
		}
	}

	defer ac.conn.Close()

	// Set to zero here to make it clear
	var total uint64 = 0

	for total < uint64(len(data)) {
		if c.conf.ReadTimeout != 0 {
			ac.conn.SetReadDeadline(time.Now().Add(c.conf.ReadTimeout))
		}
		readn, err := ac.conn.Read(data[total:])

		total += uint64(readn)
		// EOF indicates that the stream has been closed
		// Strangely, https://stackoverflow.com/questions/12741386/how-to-know-tcp-connection-is-closed-in-net-package
		// seems to claim that zero byte reads never return an error. From my experiments this is false.
		if err == io.EOF {
			return total, nil
		} else if err != io.EOF && err != nil {
			// There is a different error (such as timeout) so we must return with that error
			return total, err
		} else if readn == 0 {
			// As stated above, it is implied in some sources that reading from the closed channel
			// will return 0 bytes and no error. This does not appear to be true, but I am checking anyway
			return total, nil
		}
	}
	// Once we have filled the buffer we must check if the full message has been read
	// This is indicated by the tcp socket close handshake having occurred, which causes
	// the Read operation to return with an io.EOF error.
	// We make a buffer to hold one byte, and use it to check if there are any remaining
	// bytes to read. If there are, or the tcp socket takes too long to close,
	// then we return with an error.
	dummyBuffer := make([]byte, 1)
	// We always set a short timeout here. That way, we can check handle the case
	// where no more bytes arrive but the close handshake packets were lost.
	ac.conn.SetReadDeadline(time.Now().Add(time.Second * 5))

	readn, err := ac.conn.Read(dummyBuffer)
	if err == io.EOF || readn == 0 {
		return total, nil
	} else {
		// Too many bytes have been received (more than can be held in the buffer)
		// so we return an error message
		return total, errors.New("Buffer Full")
	}
}

func (c *Channel) Send(data []byte) (uint64, error) {

	nd := net.Dialer{
		Timeout: c.conf.DialTimeout,
	}

	ctx, cancelFn := context.WithCancel(context.Background())
	// A channel to be closed after message sending is over
	// to end the goroutine used for cancelling the dial
	doneMsg := make(chan byte)
	// If the cancel method is called for the covert channel, we want to
	// exit the dial operation. This goroutine waits for the cancel channel
	// to be closed and cancels the context if so
	go func() {
		select {
		case <-c.cancel:
			cancelFn()
		// doneDial ensures that this go routine always exits
		case <-doneMsg:
		}
	}()
	// we close the doneMsg channel to ensure that the go routine always exits
	defer close(doneMsg)

	// DialContext
	conn, err := nd.DialContext(ctx, "tcp4", (&net.TCPAddr{IP: c.conf.FriendIP[:], Port: int(c.conf.FriendReceivePort)}).String())
	if err != nil {
		return 0, err
	}

	defer conn.Close()

	if c.conf.WriteTimeout != 0 {
		conn.SetWriteDeadline(time.Now().Add(c.conf.WriteTimeout))
	}

	n, err := conn.Write(data)
	return uint64(n), err
}

// A loop to accept incoming TCP connections,
// verify if they are to the correct friend IP address,
// and if so, extract the friend port and send it to the receive method
func (c *Channel) acceptLoop() {
	var l *log.Logger = log.New(os.Stderr, "", log.Flags())
	for {
		select {
		case <-c.cancel:
			for {
				// Ensure all of the connections are closed properly before exiting
				// by emptying the acceptChan and manually closing all TCP connections
				select {
				case ac := <-c.acceptChan:
					ac.conn.Close()
				default:
					return
				}
			}
		default:
		}

		// Accept an incoming TCP connection
		// This allows this covert channel to make use of a proper 3-way TCP handshake,
		// to better obscure the covert communication that is occurring.
		conn, err := c.listener.Accept()
		if err != nil {
			continue
		}

		// Check that the TCP connection is being established from the correct
		// peer IP address
		// If not, we close the connection and continue the receive loop
		if tcpAddr, ok := conn.RemoteAddr().(*net.TCPAddr); !ok {
			conn.Close()
		} else if tcpAddr.IP == nil || bytes.Compare(tcpAddr.IP, c.conf.OriginIP[:]) != 0 {
			conn.Close()
		} else {
			// We have the correct IP address.
			// We must now record the port associated with that address to
			// restrict incoming packets to that port, and not a different port on the same machine.
			select {
			case c.acceptChan <- acceptedConn{conn: conn, friendPort: uint16(tcpAddr.Port)}:
			default:
				l.Println("Too many connections: Dropped TCP Connection")
				conn.Close()
			}
		}
	}
}
