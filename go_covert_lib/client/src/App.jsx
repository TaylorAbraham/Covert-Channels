import React, { useState, useEffect } from 'react';
import Button from 'react-bootstrap/Button';
import Form from 'react-bootstrap/Form';
import FormControl from 'react-bootstrap/FormControl';
import InputGroup from 'react-bootstrap/InputGroup';
import FormCheck from 'react-bootstrap/FormCheck';
import './styles.css';

/**
 * IMPORTANT NOTE: For styling, refer to https://getbootstrap.com/docs/4.0/utilities/position/
 */

const App = () => {
    const [textToSend, setTextToSend] = useState("");
    const [channels, setChannels] = useState([]);
    let ws;

    const sendConfig = () => {
        const cmd = JSON.stringify({OpCode : "config"});
        ws.send(cmd);
    }

    useEffect(() => {
        // Matches just the "127.0.0.1:8080" portion of the address
        const addressRegex = /[a-zA-Z0-9\.]+:[\d]+/g;
        ws = new WebSocket('ws://' + window.location.href.match(addressRegex)[0] + '/api/ws');
        ws.binaryType = 'arraybuffer';
        ws.onopen = (event) => {
            // writeResponse('Auto Connection established');
            sendConfig()
        };

        ws.onerror = (event) => {
            // writeResponse('Auto Connection error');
        };

        ws.onmessage = (event) => {
            const msg = JSON.parse(event.data);
            switch (msg.OpCode) {
                case "config":
                    console.log("### msg.Default.Channel", msg.Default.Channel);
                    // setChannels(Object.keys(msg.Default.Channel));
            }
            // if (event.data) {
            //     var data = JSON.parse(event.data);
            //     if (data.OpCode === "config") {
            //         handleConfig(data);
            //     } else {
            //         writeResponse(event.data);
            //     }
            // }
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
            <div>
                IP Addresses
        <InputGroup className="m-1 w-25">
                    <InputGroup.Prepend>
                        <InputGroup.Text className="input-text">Friend's IP</InputGroup.Text>
                    </InputGroup.Prepend>
                    <FormControl
                        placeholder="127.0.0.1"
                    />
                </InputGroup>
                <InputGroup className="m-1 w-25">
                    <InputGroup.Prepend>
                        <InputGroup.Text className="input-text">Origin IP</InputGroup.Text>
                    </InputGroup.Prepend>
                    <FormControl
                        placeholder="127.0.0.1"
                    />
                </InputGroup>
                <InputGroup className="m-1 w-25">
                    <InputGroup.Prepend>
                        <InputGroup.Text className="input-text">Bounce IP</InputGroup.Text>
                    </InputGroup.Prepend>
                    <FormControl
                        placeholder="127.0.0.1"
                    />
                </InputGroup>
            </div>
            <div>
                Ports
        <InputGroup className="m-1 w-25">
                    <InputGroup.Prepend>
                        <InputGroup.Text className="input-text">Friend's Port</InputGroup.Text>
                    </InputGroup.Prepend>
                    <FormControl
                        placeholder="3000"
                    />
                </InputGroup>
                <InputGroup className="m-1 w-25">
                    <InputGroup.Prepend>
                        <InputGroup.Text className="input-text">Origin Port</InputGroup.Text>
                    </InputGroup.Prepend>
                    <FormControl
                        placeholder="3000"
                    />
                </InputGroup>
                <InputGroup className="m-1 w-25">
                    <InputGroup.Prepend>
                        <InputGroup.Text className="input-text">Bounce Port</InputGroup.Text>
                    </InputGroup.Prepend>
                    <FormControl
                        placeholder="3000"
                    />
                </InputGroup>
            </div>
            <label className="check-container">
                <div className="check-label">Bounce</div>
                <input type="checkbox" />
                <span className="checkmark"></span>
            </label>
            <Button variant="success" className="m-1">Open Covert Channel</Button>
            <Button variant="danger">Close Covert Channel</Button>
        </div>
    );
}

export default App;
