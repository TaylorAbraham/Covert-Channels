import React from 'react';
import PropTypes from 'prop-types';
import FormControl from 'react-bootstrap/FormControl';
import InputGroup from 'react-bootstrap/InputGroup';


const IPInput = (props) => {
  const { label, value, onChange } = props;
  return (
    <InputGroup className="m-1 w-100">
      <InputGroup.Prepend>
        <InputGroup.Text className="input-text">{label}</InputGroup.Text>
      </InputGroup.Prepend>
      <FormControl
        value={value}
        onChange={onChange}
      />
    </InputGroup>
  );
};

IPInput.propTypes = {
  label: PropTypes.string.isRequired,
  value: PropTypes.string,
  onChange: PropTypes.func.isRequired,
};

IPInput.defaultProps = {
  value: '',
};

export default IPInput;
