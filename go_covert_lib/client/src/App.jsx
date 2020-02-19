import React, { useState, useEffect } from 'react';
import Navbar from 'react-bootstrap/Navbar';
import Nav from 'react-bootstrap/Nav';
import NavDropdown from 'react-bootstrap/NavDropdown';
import Button from 'react-bootstrap/Button';
import Dropdown from 'react-bootstrap/Dropdown';
import FormControl from 'react-bootstrap/FormControl';
import Spinner from 'react-bootstrap/Spinner';

import IPInput from './ui-components/IPInput';
import NumberInput from './ui-components/NumberInput';
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
    const chanConf = {};
    chanConf[channel.value] = config;
    const cmd = JSON.stringify({
      OpCode: 'open',
      Processors: processors,
      Channel: {
        Type: channel.value,
        Data: chanConf,
      },
    });
    ws.send(cmd, { binary: true });
  };

  const closeChannel = () => {
    const cmd = JSON.stringify({ OpCode: 'close' });
    ws.send(cmd, { binary: true });
  };

  const sendMessage = () => {
    const cmd = JSON.stringify({ OpCode: 'write', Message: textToSend });
    ws.send(cmd, { binary: true });
    setTextToSend('');
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
      case 'write':
        addSystemMessage('Covert message sent.');
        break;
      case 'read':
        addSystemMessage(`Covert message received: ${msg.Message}`);
        break;
      case 'error':
        addSystemMessage(`[ERROR]: ${msg.Message}`);
        break;
      default:
        console.log('ERROR: Unknown message');
        console.log('### msg', msg);
    }
  };

  useEffect(() => {
    // Matches just the "127.0.0.1:8080" portion of the address
    // const addressRegex = /[a-zA-Z0-9.]+:[\d]+/g;
    // const newWS = new WebSocket(`ws://${window.location.href.match(addressRegex)[0]}/api/ws`);
    // TODO: The line below exists for easy personal debugging
    const newWS = new WebSocket('ws://localhost:8080/api/ws');
    newWS.binaryType = 'arraybuffer';
    newWS.onopen = _e => sendInitialConfig(newWS);
    newWS.onerror = _e => console.log('UNIMPLEMENTED'); // TODO:
    newWS.onmessage = e => handleMessage(JSON.parse(e.data));
    setWS(newWS);
  }, []);

  return isLoading ? (
    <div className="spinner-container">
      <Spinner animation="border" role="status" />
    </div>
  ) : (
    <div>
      <Navbar bg="primary" variant="dark" expand="lg">
        <Navbar.Brand href="#home">Covert Client</Navbar.Brand>
        <Navbar.Toggle aria-controls="basic-navbar-nav" />
        <Navbar.Collapse id="basic-navbar-nav">
          <Nav className="mr-auto">
            <Nav.Link href="#config">Configuration</Nav.Link>
            <Nav.Link href="#msg">Messaging</Nav.Link>
            <Nav.Link href="#help">Help</Nav.Link>
          </Nav>
          <Button variant="danger">Channel Closed</Button>
        </Navbar.Collapse>
      </Navbar>
      <div className="ml-2 mb-2 mr-2 d-flex" style={{ marginTop: '70px' }}>
        {/* <h2 className="m-1">Messaging</h2>
        <FormControl
          as="textarea"
          className="w-75 m-1"
          value={textToSend}
          onChange={e => setTextToSend(e.target.value)}
        />
        <Button variant="primary" onClick={sendMessage} className="m-1">Send Message</Button>
        <br />
        <div className="m-1">System Messages</div>
        <FormControl
          as="textarea"
          className="w-75 m-1"
          value={systemMessages.join('\n')}
          readOnly
        /> */}
        <div style={{ flexGrow: 3 }}>
          <h2 className="m-1">Configuration</h2>
          <h3 className="m-1">Processors</h3>
          <Button
            variant="success"
            className="m-1 w-75"
            onClick={() => setProcessors(processors.concat({
              Type: null,
              Data: null,
            }))}
          >
            Add Processor
          </Button>
          {
            processors.map((processor, i) => (
              <div key={i.toString()}>
                <Dropdown className="m-1">
                  <Dropdown.Toggle
                    className="w-75"
                    variant="outline-primary"
                  >
                    {processor.Type || 'Select a Processor'}
                  </Dropdown.Toggle>
                  <Dropdown.Menu className="w-75">
                    {
                      Object.keys(processorList).map(p => (
                        <Dropdown.Item
                          as="option"
                          active={p === processor.Type}
                          onClick={(e) => {
                            setProcessors([
                              ...processors.slice(0, i),
                              {
                                Type: e.target.value,
                                Data: processorList,
                              },
                              ...processors.slice(i + 1, processors.length + 1),
                            ]);
                          }}
                          value={p}
                          key={p}
                        >
                          {p}
                        </Dropdown.Item>
                      ))
                    }
                  </Dropdown.Menu>
                </Dropdown>
                {processor.Data && Object.keys(processor.Data[processor.Type]).map((key) => {
                  /**
                   * EXTREME DANGER WARNING
                   * The below code involves very convoluted spread operators to massage
                   * the data to the format that GoLang expects it to be in.
                   */
                  const opt = processor.Data[processor.Type][key];
                  const props = {
                    key,
                    label: opt.Display.Name,
                    value: opt.Value,
                    onChange: e => setProcessors([
                      ...processors.slice(0, i),
                      {
                        ...processor,
                        Data: {
                          ...processor.Data,
                          [processor.Type]: {
                            ...processor.Data[processor.Type],
                            [key]: {
                              ...opt,
                              Value: e.target.value,
                            },
                          },
                        },
                      },
                      ...processors.slice(i + 1, processors.length + 1),
                    ]),
                  };
                  switch (opt.Type) {
                    case 'ipv4':
                      return (
                        <IPInput {...props} />
                      );
                    case 'i8':
                    case 'u16':
                    case 'u64':
                    case 'exactu64':
                      return (
                        <NumberInput
                          {...props}
                          onChange={e => setProcessors([
                            ...processors.slice(0, i),
                            {
                              ...processor,
                              Data: {
                                ...processor.Data,
                                [processor.Type]: {
                                  ...processor.Data[processor.Type],
                                  [key]: {
                                    ...opt,
                                    Value: parseInt(e.target.value) || 0,
                                  },
                                },
                              },
                            },
                            ...processors.slice(i + 1, processors.length + 1),
                          ])}
                        />
                      );
                    case 'bool':
                      return (
                        <Checkbox
                          {...props}
                          onChange={e => setProcessors([
                            ...processors.slice(0, i),
                            {
                              ...processor,
                              Data: {
                                ...processor.Data,
                                [processor.Type]: {
                                  ...processor.Data[processor.Type],
                                  [key]: {
                                    ...opt,
                                    Value: e.target.checked,
                                  },
                                },
                              },
                            },
                            ...processors.slice(i + 1, processors.length + 1),
                          ])}
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
              </div>
            ))
          }
          <h3 className="m-1">Channel</h3>
          <Dropdown className="m-1">
            <Dropdown.Toggle
              className="w-75"
              variant="outline-primary"
            >
              {channel.value || 'Select a Channel'}
            </Dropdown.Toggle>
            <Dropdown.Menu className="w-75">
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
              case 'i8':
              case 'u16':
              case 'u64':
              case 'exactu64':
                return (
                  <NumberInput
                    {...props}
                    onChange={e => setConfig({
                      ...config,
                      [key]: {
                        ...config[key],
                        Value: parseInt(e.target.value) || 0,
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
          <Button variant="success" onClick={openChannel} className="m-1" style={{ width: '36.25%' }}>Open Covert Channel</Button>
          <Button variant="danger" onClick={closeChannel} className="m-1" style={{ width: '36.25%' }}>Close Covert Channel</Button>
        </div>
        <div style={{ flexGrow: 1 }}>
          <div className="p-3" style={{ position: 'fixed', border: '1px solid black', height: '90%' }}>
            <h3 style={{ borderBottom: '1px solid black' }}>System Log</h3>
            <p className="">Connection to server established on port 8080.</p>
          </div>
        </div>
      </div>
    </div>
  );
};

export default App;
