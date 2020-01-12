package controller

import (
	"./channel"
	"./channel/tcpHandshake"
	"./channel/tcpNormal"
	"./channel/tcpSyn"
	"./config"
	"./processor"
	"./processor/advancedEncryptionStandard"
	"./processor/caesar"
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
		readCd configData = DefaultConfig()
		c      channel.Channel
		ps     []processor.Processor
		cconf  *channelConfig
		pconfs []processorConfig
		err    error
	)

	// Since there is always one Channel, we copy the current values
	// there so that they are filled in if some keys are ommitted
	// We don't do the same for the processors because there are a variable number of Processors
	// and we don't want to leave in extra if some have been deleted (I am not sure how the json
	// unmarshaller handles the case where the current slice is larger than the provided slice,
	// but I don't want to rely on it)
	readCd.Channel.Type = ctr.config.Channel.Type
	if err := config.CopyValueSet(&readCd.Channel.Data, ctr.config.Channel.Data, nil); err != nil {
		return nil, err
	}
	// Read in the new config data
	if err := json.Unmarshal(data, &readCd); err != nil {
		return nil, err
	}

	for i := range readCd.Processors {
		var p processor.Processor
		var pconf *processorConfig
		if p, pconf, err = ctr.retrieveProcessor(readCd.Processors[i]); err != nil {
			return nil, err
		} else {
			pconfs = append(pconfs, *pconf)
			ps = append(ps, p)
		}
	}
	if c, cconf, err = ctr.retrieveChannel(readCd.Channel); err != nil {
		return nil, err
	}
	// We only update the Processor and Channel fields, as none others should be modified
	ctr.config.Processors = pconfs
	ctr.config.Channel = *cconf

	return &Layers{processors: ps, channel: c, readClose: make(chan interface{}), readCloseDone: make(chan interface{})}, nil
}

// Retrieve the channel entity
// channelType is the type of channel
// data is the byte string of the configuration struct JSON
func (ctr *Controller) retrieveChannel(cconf channelConfig) (channel.Channel, *channelConfig, error) {
	var (
		c       channel.Channel
		newConf channelConfig
		err     error
	)
	// We must retrieve the default channel to retrieve the correct ranges
	newConf.Data = defaultChannel()
	// we create a new config and move only the new values to it
	// That way we don't override any descriptions or ranges
	newConf.Type = cconf.Type
	if err = config.CopyValueSet(&newConf.Data, cconf.Data, []string{newConf.Type}); err != nil {
		return nil, nil, err
	}

	if err = config.ValidateConfigSet(&newConf.Data); err != nil {
		return nil, nil, err
	}

	switch newConf.Type {
	case "TcpSyn":
		if c, err = tcpSyn.ToChannel(newConf.Data.TcpSyn); err != nil {
			return nil, nil, err
		}
	case "TcpHandshake":
		if c, err = tcpHandshake.ToChannel(newConf.Data.TcpHandshake); err != nil {
			return nil, nil, err
		}
	case "TcpNormal":
		if c, err = tcpNormal.ToChannel(newConf.Data.TcpNormal); err != nil {
			return nil, nil, err
		}
	default:
		err = errors.New("Invalid Channel Type")
	}
	return c, &newConf, err
}

func (ctr *Controller) retrieveProcessor(pconf processorConfig) (processor.Processor, *processorConfig, error) {
	var (
		p       processor.Processor
		newConf processorConfig
		err     error
	)
	// We must retrieve the default processor to retrieve the correct ranges
	newConf.Data = defaultProcessor()
	// we create a new config and move only the new values to it
	// That way we don't override any descriptions or ranges
	newConf.Type = pconf.Type
	if err = config.CopyValueSet(&newConf.Data, pconf.Data, []string{newConf.Type}); err != nil {
		return nil, nil, err
	}

	if err = config.ValidateConfigSet(&newConf.Data); err != nil {
		return nil, nil, err
	}

	switch newConf.Type {
	case "None":
		if p, err = none.ToProcessor(newConf.Data.None); err != nil {
			return nil, nil, err
		}
	case "Caesar":
		if p, err = caesar.ToProcessor(newConf.Data.Caesar); err != nil {
			return nil, nil, err
		}
	case "AdvancedEncryptionStandard":
		if p, err = advancedEncryptionStandard.ToProcessor(newConf.Data.AdvancedEncryptionStandard); err != nil {
			return nil, nil, err
		}
	default:
		err = errors.New("Invalid Processor Type")
	}
	return p, &newConf, err
}
