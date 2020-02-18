package controller

import (
	"encoding/json"
	"errors"

	"./channel"
	"./channel/tcpHandshake"
	"./channel/tcpNormal"
	"./channel/tcpSyn"
	"./config"
	"./processor"
	"./processor/asymmetricEncryption"
	"./processor/caesar"
	"./processor/checksum"
	"./processor/gZipCompression"
	"./processor/none"
	"./processor/symmetricEncryption"
	"./processor/zLibCompression"
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

func channelConfigCopy(currConf *channelConfig) (channelConfig, error) {
	var newConf channelConfig
	newConf.Data = defaultChannel()
	newConf.Type = currConf.Type
	if err := config.CopyValueSet(&newConf.Data, &currConf.Data, nil); err != nil {
		return newConf, err
	} else {
		return newConf, nil
	}
}

// Retrieve the layer entities that make up the covert channel
func (ctr *Controller) retrieveLayers(data []byte) (*Layers, error) {
	var (
		readCd configData = DefaultConfig()
		c      channel.Channel
		ps     []processor.Processor
		cconf  *channelConfig
		// We don't actually have to initialize these slices in go code (append does that for us)
		// but doing this ensures that null is not sent to the client
		// so that it loops properly
		pconfs []processorConfig = make([]processorConfig, 0)
		err    error
	)

	// Since there is always one Channel, we copy the current values
	// there so that they are filled in if some keys are ommitted
	// We don't do the same for the processors because there are a variable number of Processors
	// and we don't want to leave in extra if some have been deleted (I am not sure how the json
	// unmarshaller handles the case where the current slice is larger than the provided slice,
	// but I don't want to rely on it)
	// The config is used for unmarshalling the data so that empty fields are populated with their
	// current values (I have confirmed that this is how JSON unmarshal works)
	if readCd.Channel, err = channelConfigCopy(&ctr.config.Channel); err != nil {
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
	// We also populate it with the current values in the channel
	// This is effectively a copy of the current channel config
	if newConf, err = channelConfigCopy(&ctr.config.Channel); err != nil {
		return nil, nil, err
	}

	newConf.Type = cconf.Type

	// Then we populate the new config with the updated values only for the selected covert channel
	if err = config.CopyValueSet(&newConf.Data, &cconf.Data, []string{newConf.Type}); err != nil {
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
	if err = config.CopyValueSet(&newConf.Data, &pconf.Data, []string{newConf.Type}); err != nil {
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
	case "Checksum":
		if p, err = checksum.ToProcessor(newConf.Data.Checksum); err != nil {
			return nil, nil, err
		}
	case "SymmetricEncryption":
		if p, err = symmetricEncryption.ToProcessor(newConf.Data.SymmetricEncryption); err != nil {
			return nil, nil, err
		}
	case "AsymmetricEncryption":
		if p, err = asymmetricEncryption.ToProcessor(newConf.Data.AsymmetricEncryption); err != nil {
			return nil, nil, err
		}
	case "GZipCompression":
		if p, err = gZipCompression.ToProcessor(newConf.Data.GZipCompression); err != nil {
			return nil, nil, err
		}
	case "ZLibCompression":
		if p, err = zLibCompression.ToProcessor(newConf.Data.ZLibCompression); err != nil {
			return nil, nil, err
		}
	default:
		err = errors.New("Invalid Processor Type")
	}
	return p, &newConf, err
}
