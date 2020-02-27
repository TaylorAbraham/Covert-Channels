import React from 'react';
import PropTypes from 'prop-types';
import FormControl from 'react-bootstrap/FormControl';
import Button from 'react-bootstrap/Button';

const MessagingScreen = (props) => {
  const {
    textToSend,
    setTextToSend,
    covertMessages,
    sendMessage,
  } = props;
  return (
    <div className="cc-messaging m-2">
      <h2 className="m-1">Messaging</h2>
      <FormControl
        as="textarea"
        className="cc-messaging__input w-75 m-1"
        value={textToSend}
        onChange={e => setTextToSend(e.target.value)}
      />
      <Button
        variant="primary"
        onClick={sendMessage}
        className="cc-messaging__send m-1"
        disabled={!textToSend}
      >
        Send Message
      </Button>
      <br />
      <div className="m-1">Incoming Messages</div>
      <FormControl
        as="textarea"
        className="cc-messaging__output w-75 m-1"
        value={covertMessages.join('\n')}
        readOnly
      />
    </div>
  );
};

MessagingScreen.propTypes = {
  textToSend: PropTypes.string.isRequired,
  setTextToSend: PropTypes.func.isRequired,
  covertMessages: PropTypes.array.isRequired,
  sendMessage: PropTypes.func.isRequired,
};

export default MessagingScreen;
