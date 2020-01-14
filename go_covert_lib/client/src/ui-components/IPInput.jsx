import React from 'react';
import PropTypes from 'prop-types';
import FormControl from 'react-bootstrap/FormControl';
import InputGroup from 'react-bootstrap/InputGroup';


const IPInput = (props) => {
  const { label, value } = props;
  return (
    <InputGroup className="m-1 w-25">
      <InputGroup.Prepend>
        <InputGroup.Text className="input-text">{label}</InputGroup.Text>
      </InputGroup.Prepend>
      <FormControl
        value={value}
      />
    </InputGroup>
  );
};

IPInput.propTypes = {
  label: PropTypes.string.isRequired,
  value: PropTypes.string,
};

IPInput.defaultProps = {
  value: '',
};

export default IPInput;
