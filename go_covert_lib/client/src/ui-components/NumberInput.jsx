import React from 'react';
import PropTypes from 'prop-types';
import FormControl from 'react-bootstrap/FormControl';
import InputGroup from 'react-bootstrap/InputGroup';

const NumberInput = (props) => {
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

NumberInput.propTypes = {
  label: PropTypes.string.isRequired,
  value: PropTypes.number,
  onChange: PropTypes.func.isRequired,
};

NumberInput.defaultProps = {
  value: 0,
};

export default NumberInput;
