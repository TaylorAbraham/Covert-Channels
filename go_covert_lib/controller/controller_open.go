package controller

import (
  "encoding/json"
  "errors"
  "./config"
  "./channel"
  "./channel/ipv4TCP"
  "./processor"
  "./processor/none"
)

func (ctr *Controller) handleOpen(data []byte) error {
  if l, err := ctr.retrieveLayers(data); err == nil {
    ctr.layers = l
    return nil
  } else {
    return err
  }
}

func (ctr *Controller) retrieveLayers(data []byte) (*Layers, error) {
  var (
    ct configType
    c channel.Channel
    p processor.Processor
    err error
  )
  if err := json.Unmarshal(data, &ct); err != nil {
    return nil, err
  }

  if p, err = ctr.retrieveProcessor(ct.ProcessorType, data); err != nil {
    return nil, err
  }
  if c, err = ctr.retrieveChannel(ct.ChannelType, data); err != nil {
    return nil, err
  }

  ctr.config.ProcessorType = ct.ProcessorType
  ctr.config.ChannelType = ct.ChannelType

  return &Layers{ processor : p, channel : c, readClose : make(chan interface{}), readCloseDone : make(chan interface{})}, nil
}

func (ctr *Controller) retrieveChannel(channelType string, data []byte) (channel.Channel, error) {
  var (
    c channel.Channel
    err error
  )

  switch channelType {
  case "Ipv4TCP":
    var tempconf configData = DefaultConfig()
    var itConf ipv4TCP.ConfigClient = ipv4TCP.GetDefault()
    var ipCh *ipv4TCP.Channel
    if err = unmarshalCopyValidate(data, &tempconf,
      &ctr.config.Channel.Ipv4TCP, &tempconf.Channel.Ipv4TCP, &itConf,
      func () error {var err error; ipCh, err = ipv4TCP.ToChannel(itConf); return err}); err != nil {
      if ipCh != nil {
        ipCh.Close();
      }
    } else {
      c = ipCh
    }
  default:
    err = errors.New("Invalid Channel Type")
  }
  return c, err
}

func (ctr *Controller) retrieveProcessor(processorType string, data []byte) (processor.Processor, error) {
  var (
    p processor.Processor
    err error
  )

  switch processorType {
  case "None":
    var tempconf configData = DefaultConfig()
    var noneConf none.ConfigClient = none.GetDefault()
    err = unmarshalCopyValidate(data, &tempconf,
     &ctr.config.Processor.None, &tempconf.Processor.None, &noneConf,
     func () error {var err error; p, err = none.ToProcessor(noneConf); return err})
  default:
    err = errors.New("Invalid Processor Type")
  }
  return p, err
}

// Five steps:
//  copy originalItem to tempItem. This way the original is not changed if we find an error when validating (tempItem is the specific config found in temp)
//  unmarshal into temp. This way only the temp is updated during unmarshalling
//  copy tempItem into newItem to preserve the correct range values (i.e. the UI can't overwrite them)
//  Execute the function f to create the channel or processor
//  validate the newItem (which has been updated with the new values)
func unmarshalCopyValidate(data []byte, temp interface{}, originalItem interface{}, tempItem interface{}, newItem interface{}, f func() error) error {
  if err := config.CopyValue(tempItem, originalItem); err != nil {
    return err
  }
  if err := json.Unmarshal(data, temp); err != nil {
    return err
  }
  // We must copy to ensure that the JSON does not
  // overwrite any of the range values
  if err := config.CopyValue(newItem, tempItem); err != nil {
    return err
  }
  if err := config.Validate(newItem); err != nil {
    return err
  }

  if err := f(); err != nil {
    return err
  }

  if err := config.CopyValue(originalItem, newItem); err != nil {
    return err
  }
  return nil
}