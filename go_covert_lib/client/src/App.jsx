import React, { useState, useEffect } from 'react';
import Button from 'react-bootstrap/Button';
import Dropdown from 'react-bootstrap/Dropdown';
import FormControl from 'react-bootstrap/FormControl';
import Spinner from 'react-bootstrap/Spinner';

import IPInput from './ui-components/IPInput';
import PortInput from './ui-components/PortInput';
import './styles.css';
import Checkbox from './ui-components/Checkbox';
import Select from './ui-components/Select';

/**
 * IMPORTANT NOTE: For styling, refer to https://getbootstrap.com/docs/4.0/utilities/position/
 */
const App = () => {
  const [textToSend, setTextToSend] = useState('');
  const [processorList, setProcessorList] = useState([]);
  const [processors, setProcessors] = useState([]);
  const [channelList, setChannelList] = useState([]);
  const [channel, setChannel] = useState({});
  const [config, setConfig] = useState({});
  const [isLoading, setLoading] = useState(true);
  const [ws, setWS] = useState(null);
  const [systemMessages, setSystemMessages] = useState([]);

  const sendInitialConfig = (localWS) => {
    const cmd = JSON.stringify({ OpCode: 'config' });
    localWS.send(cmd, { binary: true });
  };

  const addSystemMessage = (newMsg) => {
    setSystemMessages(sm => sm.concat(newMsg));
  };

  const openChannel = () => {
    const cmd = JSON.stringify({
      OpCode: 'open',
      Processors: config.processors,
      Channel: {
        Type: channel.value,
        Data: config,
      },
    });
    ws.send(cmd, { binary: true });
  };

  const closeChannel = () => {
    const cmd = JSON.stringify({ OpCode: 'close' });
    ws.send(cmd, { binary: true });
  };

  const handleMessage = (msg) => {
    switch (msg.OpCode) {
      case 'config':
        setChannelList(msg.Default.Channel);
        setProcessorList(msg.Default.Processor);
        addSystemMessage('Connection to server established.');
        setLoading(false);
        break;
      case 'open':
        addSystemMessage('Covert channel successfully opened.');
        break;
      case 'close':
        addSystemMessage('Covert channel closed.');
        break;
      default:
        console.log('ERROR: Unknown message');
        console.log("### msg", msg);
          // TODO:
    }
  };

  useEffect(() => {
    // Matches just the "127.0.0.1:8080" portion of the address
    // const addressRegex = /[a-zA-Z0-9.]+:[\d]+/g;
    // const ws = new WebSocket(`ws://${window.location.href.match(addressRegex)[0]}/api/ws`);
    // TODO: The line below exists for easy personal debugging
    const newWS = new WebSocket('ws://localhost:8080/api/ws');
    newWS.binaryType = 'arraybuffer';
    newWS.onopen = _e => sendInitialConfig(newWS);
    newWS.onerror = _e => console.log('UNIMPLEMENTED'); // TODO:
    newWS.onmessage = e => handleMessage(JSON.parse(e.data));
    setWS(newWS);
  }, []);

  console.log("### config", config);

  return isLoading ? (
    <div className="spinner-container">
      <Spinner animation="border" role="status" />
    </div>
  ) : (
    <div className="m-2">
      <h2 className="m-1">Messaging</h2>
      <FormControl
        as="textarea"
        className="w-25 m-1"
        value={textToSend}
        onChange={e => setTextToSend(e.target.value)}
      />
      <Button variant="primary" className="m-1">Send Message</Button>
      <br />
      <div className="m-1">Incoming Messages</div>
      <FormControl
        as="textarea"
        className="w-50 m-1"
        readOnly
      />
      <div className="m-1">System Messages</div>
      <FormControl
        as="textarea"
        className="w-50 m-1"
        value={systemMessages.join('\n')}
        readOnly
      />
      <h2 className="m-1 mt-5">Configuration</h2>
      <Dropdown className="m-1">
        <Dropdown.Toggle
          className="w-25"
          variant="outline-primary"
        >
          {channel.value || 'Select a Channel'}
        </Dropdown.Toggle>
        <Dropdown.Menu className="w-25">
          {
            Object.keys(channelList).map(chan => (
              <Dropdown.Item
                as="option"
                active={chan === channel.value}
                onClick={(e) => {
                  setChannel({
                    value: e.target.value,
                    properties: channelList[chan],
                  });
                  setConfig(channelList[chan]);
                }}
                value={chan}
                key={chan}
              >
                {chan}
              </Dropdown.Item>
            ))
          }
        </Dropdown.Menu>
      </Dropdown>
      {Object.keys(config).map((key) => {
        const opt = config[key];
        const props = {
          key,
          label: opt.Display.Name,
          value: opt.Value,
          onChange: e => setConfig({
            ...config,
            [key]: {
              ...config[key],
              Value: e.target.value,
            },
          }),
        };
        switch (opt.Type) {
          case 'ipv4':
            return (
              <IPInput {...props} />
            );
          case 'u16':
          case 'u64':
            return (
              <PortInput
                {...props}
                onChange={e => setConfig({
                  ...config,
                  [key]: {
                    ...config[key],
                    Value: parseInt(e.target.value),
                  },
                })}
              />
            );
          case 'bool':
            return (
              <Checkbox
                {...props}
                onChange={e => setConfig({
                  ...config,
                  [key]: {
                    ...config[key],
                    Value: e.target.checked,
                  },
                })}
              />
            );
          case 'select':
            return (
              <Select
                {...props}
                items={opt.Range}
              />
            );
          default:
            return (<div key={key}>UNIMPLEMENTED</div>);
        }
      })}
      <Button variant="success" onClick={() => openChannel()} className="m-1 w-25">Open Covert Channel</Button>
      <Button variant="danger" onClick={() => closeChannel()} className="m-1 w-25">Close Covert Channel</Button>
    </div>
  );
};

export default App;
