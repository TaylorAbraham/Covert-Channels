import React, { useState, useEffect } from 'react';
import Button from 'react-bootstrap/Button';
import Dropdown from 'react-bootstrap/Dropdown';
import FormControl from 'react-bootstrap/FormControl';
import './styles.css';

// Matches just the "127.0.0.1:8080" portion of the address
const addressRegex = /[a-zA-Z0-9.]+:[\d]+/g;
const ws = new WebSocket(`ws://${window.location.href.match(addressRegex)[0]}/api/ws`);
// TODO: The below exists for easy personal debugging
// const ws = new WebSocket('ws://localhost:8080/api/ws');

/**
 * IMPORTANT NOTE: For styling, refer to https://getbootstrap.com/docs/4.0/utilities/position/
 */
const App = () => {
  const [textToSend, setTextToSend] = useState('');
  const [processors, setProcessors] = useState([]);
  const [processor, setProcessor] = useState({});
  const [channels, setChannels] = useState([]);
  const [channel, setChannel] = useState({});

  const sendConfig = () => {
    const cmd = JSON.stringify({ OpCode: 'config' });
    ws.send(cmd);
  };

  useEffect(() => {
    ws.binaryType = 'arraybuffer';
    ws.onopen = (event) => {
      sendConfig();
    };

    ws.onerror = (event) => {
      // TODO:
    };

    ws.onmessage = (event) => {
      const msg = JSON.parse(event.data);
      switch (msg.OpCode) {
        case 'config':
          setChannels(msg.Default.Channel);
          setProcessors(msg.Default.Processor);
          break;
        default:
            // TODO:
      }
    };
  }, []);

  return (
    <div className="App m-2">
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
        className="w-25 m-1"
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
            Object.keys(channels).map(chan => (
              <Dropdown.Item
                as="option"
                active={chan === channel.value}
                onClick={e => setChannel({
                  value: e.target.value,
                  properties: channels[chan],
                })}
                value={chan}
                key={chan}
              >
                {chan}
              </Dropdown.Item>
            ))
          }
        </Dropdown.Menu>
      </Dropdown>
      <Button variant="success" className="m-1 w-25">Open Covert Channel</Button>
      <Button variant="danger" className="m-1 w-25">Close Covert Channel</Button>
    </div>
  );
};

export default App;
