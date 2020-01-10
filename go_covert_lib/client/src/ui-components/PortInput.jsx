import React from 'react';
import PropTypes from 'prop-types';
import FormControl from 'react-bootstrap/FormControl';
import InputGroup from 'react-bootstrap/InputGroup';


const PortInput = (props) => {
  const { label } = props;
  return (
    <InputGroup className="m-1 w-25">
      <InputGroup.Prepend>
        <InputGroup.Text className="input-text">{label}</InputGroup.Text>
      </InputGroup.Prepend>
      <FormControl
        placeholder="8080"
      />
    </InputGroup>
  );
};

PortInput.propTypes = {
  label: PropTypes.string.isRequired,
};

export default PortInput;
