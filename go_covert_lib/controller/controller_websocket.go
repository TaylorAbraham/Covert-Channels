package controller

import (
	"errors"
	"github.com/gorilla/websocket"
	"log"
	"net/http"
	"time"
)

// The HTTP handler function for initializing and running the websocket
func (ctr *Controller) HandleFunc(w http.ResponseWriter, r *http.Request) {
	r.Header.Del("Origin")
	ctr.clientLock.Lock()

	select {
	case <-ctr.clientStop:
		ctr.clientLock.Unlock()
		return
	default:
	}
	ctr.waitGroup.Add(1)

	ws, err := ctr.upgrader.Upgrade(w, r, nil)
	if err != nil {
		ctr.clientLock.Unlock()
		log.Print(err)
		return
	}

	ctr.clients[ws] = true
	ctr.clientLock.Unlock()

	defer func() {
		ctr.clientLock.Lock()
		delete(ctr.clients, ws)
		ctr.clientLock.Unlock()
		// Make sure we close the  connection when the function returns
		ws.Close()
		ctr.waitGroup.Done()
	}()

loop:
	for {
		_, data, err := ws.ReadMessage()
		if err == nil {
			select {
			case ctr.wsRecv <- data:
			case <-ctr.clientStop:
				break loop
			case <-time.After(time.Second):
			}
		} else {
			log.Println("Websocket read error: " + err.Error())
			break loop
		}
	}
}

// A loop for processing incomming messages from the client
func (ctr *Controller) webReceiveLoop() {
	defer close(ctr.doneWsRecv)

loop:
	for {
		select {
		case <-ctr.recvStop:
			break loop
		case data := <-ctr.wsRecv:
			ctr.wsSend <- ctr.handleMessage(data)
		}
	}
}

// A loop for broadcasting outgoing messages along all websockets
func (ctr *Controller) webSendLoop() {
	defer close(ctr.doneWsSend)

loop:
	for {
		select {
		case <-ctr.sendStop:
			break loop
		case data := <-ctr.wsSend:
			ctr.clientLock.Lock()
			for ws, _ := range ctr.clients {
				if err := ws.WriteMessage(websocket.TextMessage, data); err != nil {
					log.Println("Websocket write error: " + err.Error())
				}
			}
			ctr.clientLock.Unlock()
		}
	}
}

// Shutdown the websocket and all send and receive loops
func (ctr *Controller) webShutdown() error {
	var err error
	ctr.clientLock.Lock()
	close(ctr.clientStop)
	for c := range ctr.clients {
		c.Close()
	}

	ctr.clients = make(map[*websocket.Conn]bool)
	ctr.clientLock.Unlock()

	//Wait until all HandleConnection functions have completed
	ctr.waitGroup.Wait()

	close(ctr.recvStop)

	select {
	case <-ctr.doneWsRecv:
	case <-time.After(time.Second * 5):
		err = errors.New("Failed to stop recv loop")
	}

	close(ctr.sendStop)

	select {
	case <-ctr.doneWsSend:
	case <-time.After(time.Second * 5):
		err = errors.New("Failed to stop recv loop")
	}
	return err
}
