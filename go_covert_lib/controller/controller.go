package controller

import (
  "encoding/json"
  "errors"
  "strconv"
  "time"
  "github.com/gorilla/websocket"
  "./config"
  "./channel/ipv4TCP"
  "./processor/none"
)

func CreateController() (*Controller, error) {
  var ctr *Controller = &Controller{
    config : DefaultConfig(),
    clients    : make(map[*websocket.Conn]bool),
    clientStop : make(chan interface{}),
    recvStop   : make(chan interface{}),
    sendStop   : make(chan interface{}),
    doneWsSend : make(chan interface{}),
    doneWsRecv : make(chan interface{}),
    wsSend     : make(chan []byte),
    wsRecv     : make(chan []byte),
  }
  if err := config.ValidateConfigSet(ctr.config.Processor); err != nil {
    return nil, err
  }
  if err := config.ValidateConfigSet(ctr.config.Channel); err != nil {
    return nil, err
  }
  go ctr.webReceiveLoop()
  go ctr.webSendLoop()
  return ctr, nil
}

func DefaultConfig() configData {
  return configData{
    OpCode : "config",
    ChannelType : "Ipv4TCP",
    ProcessorType : "None",
    Processor : processorData {
      None : none.GetDefault(),
    },
    Channel : channelData {
      Ipv4TCP : ipv4TCP.GetDefault(),
    },
  }
}

func (ctr *Controller) handleMessage(data []byte) []byte {
  var cmd command
  if err := json.Unmarshal(data, &cmd); err != nil {
    return toMessage("error", "Unable to read command: " + err.Error())
  }

  switch cmd.OpCode {
  case "open":
    // Close a channel if it is already open
    if err := ctr.handleClose(); err != nil {
      return toMessage("error", "Unable to close channel: " + err.Error())
    } else if err := ctr.handleOpen(data); err != nil {
      return toMessage("error", "Unable to open channel: " + err.Error())
    } else {
      go ctr.readLoop()
      return toMessage("open", "Open success")
    }
  case "close":
    if err := ctr.handleClose(); err != nil {
      return toMessage("error", "Unable to close channel: " + err.Error())
    } else {
      return toMessage("close", "Close success")
    }
  case "write":
    if err := ctr.handleWrite(data); err != nil {
      return toMessage("error", "Unable to write to channel: " + err.Error())
    } else {
      return toMessage("write", "Message write success")
    }
  case "config":
    if data, err := ctr.handleConfig(); err != nil {
      return toMessage("error", "Could not encode config: " + err.Error())
    } else {
      return data
    }
  default:
    return toMessage("error", "Unknown operation code")
  }
}

func toMessage(opcode string, data string) []byte {
  var mt messageType
  mt.OpCode = opcode
  mt.Message = data
  if data, err := json.Marshal(mt); err != nil {
    return []byte("{\"OpCode\" : \"error\", \"Message\" : \"Marshal Error\" }")
  } else {
    return data
  }
}

func (ctr *Controller) handleConfig() ([]byte, error) {
  if data, err := json.Marshal(ctr.config); err != nil {
    return nil, err
  } else {
    return data, nil
  }
}

func (ctr *Controller) handleWrite(data []byte) error {
  var mt messageType
  if err := json.Unmarshal(data, &mt); err != nil {
    return err
  }
  if ctr.layers == nil {
    return errors.New("Channel closed")
  }
  if b, err := ctr.layers.processor.Process([]byte(mt.Message)); err != nil  {
    return errors.New("Unable to process outgoing message: " + err.Error())
  } else {
    if n, err := ctr.layers.channel.Send(b, nil); err != nil  {
      return errors.New("Write fail: Wrote " + strconv.FormatUint(n, 10) + "bytes out of " + strconv.FormatUint(uint64(len(b)), 10)  + ": " + err.Error())
    } else {
      return nil
    }
  }
}

func (ctr *Controller) handleRead() ([]byte, error) {

  var buffer [1024]byte

  if n, err := ctr.layers.channel.Receive(buffer[:], nil); err != nil {
    return nil, errors.New("Read fail: Read " + strconv.FormatUint(n, 10) + " bytes out of " + strconv.FormatUint(uint64(len(buffer)), 10)  + " available bytes: " + err.Error())
  } else {
    if data, err := ctr.layers.processor.Unprocess(buffer[:n]); err != nil  {
      return nil, errors.New("Unable to unprocess incoming message: " + err.Error())
    } else {
      return data, nil
    }
  }
}

func (ctr *Controller) readLoop() {
  loop:
  for {
    select {
    case <-ctr.layers.readClose:
      close(ctr.layers.readCloseDone)
      break loop
    default:
      data, err := ctr.handleRead()
      if err != nil {
        ctr.wsSend <- toMessage("error", err.Error())
        // If there has been a read error wait
        // to avoid a constant stream of data
        // to the UI
        time.Sleep(time.Second)
      } else {
        ctr.wsSend <- toMessage("read", string(data))
      }
    }
  }
}

func (ctr *Controller) handleClose() error {
  var err error
  if ctr.layers != nil {
    err = ctr.layers.channel.Close()
    // We must wait to ensure that the read loop is complete
    close(ctr.layers.readClose)
    <-ctr.layers.readCloseDone
    ctr.layers = nil
  }
  return err
}

func (ctr *Controller) Shutdown() error {
  return ctr.webShutdown()
}
