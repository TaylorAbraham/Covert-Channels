import React, { useState } from 'react';
import 'bootstrap/dist/css/bootstrap.css';
/**
 * IMPORTANT NOTE: For styling, refer to https://getbootstrap.com/docs/4.0/utilities
 */

const App: React.FC = () => {
  const [textToSend, setTextToSend] = useState("");

  return (
    <div className="App m-2">
      <textarea
        className="form-control w-25 m-1"
        value={textToSend}
        onChange={e => setTextToSend(e.target.value)}
      />
      <button type="button" className="btn btn-primary m-1">Send Message</button>
      <br />
      <div className="m-1">Incoming Messages</div>
      <textarea
        className="form-control w-25 m-1"
        readOnly
      />
      <h2 className="m-1">Configuration</h2>
      <button type="button" className="btn btn-primary m-1">Open Covert Channel</button>
      <button type="button" className="btn btn-primary m-1">Close Covert Channel</button>
    </div>
  );
}

export default App;
