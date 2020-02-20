import React from 'react';
import PropTypes from 'prop-types';

import './Console.scss';
import Button from 'react-bootstrap/Button';

const Console = (props) => {
  const { messages, consoleIsVisible, toggleConsoleIsVisible } = props;

  return (
    <div className="cc-console">
      <div>
        <h2 className="cc-console__header p-2">
          <Button
            className={`cc-console__toggle-btn ${!consoleIsVisible ? 'cc-console__toggle-btn--reverse' : ''}`}
            onClick={toggleConsoleIsVisible}
          >
            <p>&gt;</p>
          </Button>
          {' System Messages'}
        </h2>
      </div>
      <div className="cc-console__messages">
        {messages.map(msg => (
          <div key={msg} className="pt-1 pl-1 pr-1">
            {msg}
          </div>
        ))}
      </div>
    </div>
  );
};

Console.propTypes = {
  messages: PropTypes.arrayOf(PropTypes.string).isRequired,
  consoleIsVisible: PropTypes.bool.isRequired,
  toggleConsoleIsVisible: PropTypes.func.isRequired,
};

export default Console;
