import React, { useState, useEffect } from 'react';
import PropTypes from 'prop-types';

import './Console.scss';

const Console = (props) => {
  const { messages } = props;

  return (
    <div className="cc-console">
      <h2 className="cc-console__header p-2">System Messages</h2>
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
};

export default Console;
