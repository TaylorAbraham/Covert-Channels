import React, { useState } from 'react';
import Button from 'react-bootstrap/Button';
import FormControl from 'react-bootstrap/FormControl';

/**
 * IMPORTANT NOTE: For styling, refer to https://getbootstrap.com/docs/4.0/utilities
 */

const App: React.FC = () => {
  const [textToSend, setTextToSend] = useState("");

  return (
    <div className="App m-2">
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
      <h2 className="m-1">Configuration</h2>
      <Button variant="success" className="m-1">Open Covert Channel</Button>
      <Button variant="danger">Close Covert Channel</Button>
    </div>
  );
}

export default App;
