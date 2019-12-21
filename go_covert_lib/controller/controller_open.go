package controller

import (
	"./channel"
	"./channel/ipv4TCP"
	"./config"
	"./processor"
	"./processor/none"
	"encoding/json"
	"errors"
)

// Function for opening a covert channel
// Input is the byte string representing a JSON object with the configuration for the channel
func (ctr *Controller) handleOpen(data []byte) error {
	if l, err := ctr.retrieveLayers(data); err == nil {
		ctr.layers = l
		return nil
	} else {
		return err
	}
}

// Retrieve the layer entities that make up the covert channel
func (ctr *Controller) retrieveLayers(data []byte) (*Layers, error) {
	var (
		ct  configType
		c   channel.Channel
		p   processor.Processor
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

	return &Layers{processor: p, channel: c, readClose: make(chan interface{}), readCloseDone: make(chan interface{})}, nil
}

// Retrieve the channel entity
// channelType is the type of channel
// data is the byte string of the configuration struct JSON
func (ctr *Controller) retrieveChannel(channelType string, data []byte) (channel.Channel, error) {
	var (
		c        channel.Channel
		err      error
		tempconf configData = DefaultConfig()
	)

	switch channelType {
	case "Ipv4TCP":
		var (
			itConf ipv4TCP.ConfigClient = ipv4TCP.GetDefault()
			ipCh   *ipv4TCP.Channel
		)
		if err = unmarshalCopyValidate(data, &tempconf,
			&ctr.config.Channel.Ipv4TCP, &tempconf.Channel.Ipv4TCP, &itConf,
			func() error { var err error; ipCh, err = ipv4TCP.ToChannel(itConf); return err }); err != nil {
			if ipCh != nil {
				ipCh.Close()
			}
		} else {
			c = ipCh
		}
	default:
		err = errors.New("Invalid Channel Type")
	}
	return c, err
}

// Retrieve the processor entity
// processorType is the type of processor
// data is the byte string of the configuration struct JSON
func (ctr *Controller) retrieveProcessor(processorType string, data []byte) (processor.Processor, error) {
	var (
		p   processor.Processor
		err error
	)

	switch processorType {
	case "None":
		var tempconf configData = DefaultConfig()
		var noneConf none.ConfigClient = none.GetDefault()
		err = unmarshalCopyValidate(data, &tempconf,
			&ctr.config.Processor.None, &tempconf.Processor.None, &noneConf,
			func() error { var err error; p, err = none.ToProcessor(noneConf); return err })
	default:
		err = errors.New("Invalid Processor Type")
	}
	return p, err
}

// The following function simplifies unarshalling, validating, and copying the new config, as well as executing a function to create the new channel
// originalItem, tempItem, and newItem must all be the same config type
// tempItem must be an instance of a config found in the temp instance
func unmarshalCopyValidate(data []byte, temp interface{}, originalItem interface{}, tempItem interface{}, newItem interface{}, f func() error) error {
	//  copy values from originalItem to tempItem. This way the original is not changed if we find an error when validating and if the incoming configuration
	//  lacks parameters they are replaced with the current values.
	// (tempItem is the specific config found in temp)
	if err := config.CopyValue(tempItem, originalItem); err != nil {
		return err
	}
	//  unmarshal into temp. This way only the temp is updated during unmarshalling
	if err := json.Unmarshal(data, temp); err != nil {
		return err
	}
	// copy tempItem into newItem to restore the correct range values (i.e. the UI can't overwrite them)
	// newItem must have been preinitialized
	if err := config.CopyValue(newItem, tempItem); err != nil {
		return err
	}
	// validate the newItem (which has been updated with the new values)
	if err := config.Validate(newItem); err != nil {
		return err
	}
	// Execute the function f to create the channel or processor
	if err := f(); err != nil {
		return err
	}
	// If all validation passes, copy newItem values to originalItem values
	if err := config.CopyValue(originalItem, newItem); err != nil {
		return err
	}
	return nil
}
