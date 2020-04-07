import React, { useState, useEffect } from 'react';
import Navbar from 'react-bootstrap/Navbar';
import Nav from 'react-bootstrap/Nav';
import Button from 'react-bootstrap/Button';
import Spinner from 'react-bootstrap/Spinner';

import ConfigScreen from './screens/ConfigScreen';
import Console from './screens/Console';
import './styles.scss';
import MessagingScreen from './screens/MessagingScreen';
import HelpScreen from './screens/HelpScreen';

const Screens = Object.freeze({
  CONFIG: 'config',
  MSG: 'msg',
  HELP: 'help',
});

const getTimestamp = () => {
  const d = new Date();
  const hr = d.getHours().toString().padStart(2, '0');
  const min = d.getMinutes().toString().padStart(2, '0');
  const sec = d.getSeconds().toString().padStart(2, '0');
  const ms = d.getMilliseconds().toString().padStart(3, '0');
  return `${hr}:${min}:${sec}.${ms}`;
};

/**
 * IMPORTANT NOTE: For styling, refer to https://getbootstrap.com/docs/4.0/utilities/position/
 */
const App = () => {
  const [textToSend, setTextToSend] = useState('');
  const [processorList, setProcessorList] = useState([]);
  const [processors, setProcessors] = useState([]);
  const [channelList, setChannelList] = useState([]);
  const [channel, setChannel] = useState({});
  const [channelIsOpen, setChannelIsOpen] = useState(false);
  const [consoleIsVisible, setConsoleIsVisible] = useState(true);
  const [config, setConfig] = useState({});
  const [isLoading, setLoading] = useState(true);
  const [ws, setWS] = useState(null);
  const [systemMessages, setSystemMessages] = useState([]);
  const [covertMessages, setCovertMessages] = useState([]);
  const [screen, setScreen] = useState(Screens.CONFIG);

  const sendInitialConfig = (localWS) => {
    const cmd = JSON.stringify({ OpCode: 'config' });
    localWS.send(cmd, { binary: true });
  };

  const addSystemMessage = (newMsg) => {
    setSystemMessages(sm => sm.concat(`[${getTimestamp()}] ${newMsg}`));
  };

  const addCovertMessage = (newMsg) => {
    setCovertMessages(cm => cm.concat(`[${getTimestamp()}] ${newMsg}`));
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
        setChannelIsOpen(true);
        break;
      case 'close':
        setChannelIsOpen(false);
        addSystemMessage('Covert channel closed.');
        break;
      case 'write':
        addSystemMessage('Covert message sent.');
        break;
      case 'read':
        addSystemMessage('Covert message received.');
        addCovertMessage(msg.Message);
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
    const addressRegex = /[a-zA-Z0-9.]+:[\d]+/g;
    const newWS = new WebSocket(`ws://${window.location.href.match(addressRegex)[0]}/api/ws`);
    // TODO: The line below exists for easy personal debugging
    // const newWS = new WebSocket('ws://localhost:8080/api/ws');
    newWS.binaryType = 'arraybuffer';
    newWS.onopen = _e => sendInitialConfig(newWS);
    newWS.onerror = e => addSystemMessage(e);
    newWS.onmessage = e => handleMessage(JSON.parse(e.data));
    setWS(newWS);

    if (!window.File || !window.FileReader || !window.FileList || !window.Blob) {
      addSystemMessage('[WARNING] Your browser does not support file operations. Saving/loading configurations will not work.');
    }
  }, []);

  return isLoading ? (
    <div className="spinner-container">
      <Spinner animation="border" role="status" />
    </div>
  ) : (
    <div>
      <Navbar className="cc-navbar" bg="primary" variant="dark" expand="lg">
        <Navbar.Brand href="#config">Covert Client</Navbar.Brand>
        <Navbar.Toggle aria-controls="basic-navbar-nav" />
        <Navbar.Collapse id="basic-navbar-nav">
          <Nav activeKey={`#${screen}`} className="mr-auto">
            <Nav.Link href="#config" onClick={() => setScreen(Screens.CONFIG)}>Configuration</Nav.Link>
            <Nav.Link href="#msg" onClick={() => setScreen(Screens.MSG)}>Messaging</Nav.Link>
            <Nav.Link href="#help" onClick={() => setScreen(Screens.HELP)}>Help</Nav.Link>
          </Nav>
          {channelIsOpen ? (
            <Button variant="success" disabled style={{ opacity: '100%' }}>Channel Open</Button>
          ) : (
            <Button variant="danger" disabled style={{ opacity: '100%' }}>Channel Closed</Button>
          )}
        </Navbar.Collapse>
      </Navbar>
      <div className="cc-content">
        <div className={`cc-content__screen ${!consoleIsVisible ? 'cc-content__screen--expand' : ''}`}>
          {screen === Screens.CONFIG ? (
            <ConfigScreen
              openChannel={openChannel}
              closeChannel={closeChannel}
              config={config}
              setConfig={setConfig}
              processorList={processorList}
              processors={processors}
              setProcessors={setProcessors}
              channelList={channelList}
              channel={channel}
              setChannel={setChannel}
              channelIsOpen={channelIsOpen}
              addSystemMessage={addSystemMessage}
            />
          ) : (screen === Screens.MSG) ? (
            <MessagingScreen
              textToSend={textToSend}
              setTextToSend={setTextToSend}
              covertMessages={covertMessages}
              sendMessage={sendMessage}
            />
          ) : (screen === Screens.HELP) ? (
            <HelpScreen />
          ) : (
            <div>UNIMPLEMENTED</div>
          )}
        </div>
        <div className={`cc-content__console ${!consoleIsVisible ? 'cc-content__console--hidden' : ''}`}>
          <Console
            messages={systemMessages}
            consoleIsVisible={consoleIsVisible}
            toggleConsoleIsVisible={() => setConsoleIsVisible(!consoleIsVisible)}
          />
        </div>
      </div>
    </div>
  );
};

export default App;
