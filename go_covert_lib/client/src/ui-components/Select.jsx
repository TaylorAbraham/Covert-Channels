import React from 'react';
import PropTypes from 'prop-types';
import Dropdown from 'react-bootstrap/Dropdown';
import InputGroup from 'react-bootstrap/InputGroup';

const Select = (props) => {
  const {
    label,
    items,
    value,
    onChange,
  } = props;
  return (
    <InputGroup className="d-flex m-1 w-25">
      <InputGroup.Prepend>
        <InputGroup.Text className="input-text">{label}</InputGroup.Text>
      </InputGroup.Prepend>
      <Dropdown style={{ flexGrow: 1 }}>
        <Dropdown.Toggle
          variant="outline-primary"
          className="w-100"
        >
          {value}
        </Dropdown.Toggle>
        <Dropdown.Menu className="w-100">
          {
            items.map(item => (
              <Dropdown.Item
                as="option"
                active={value === item}
                onClick={onChange}
                value={item}
                key={item}
              >
                {item}
              </Dropdown.Item>
            ))
          }
        </Dropdown.Menu>
      </Dropdown>
    </InputGroup>
  );
};

Select.propTypes = {
  label: PropTypes.string.isRequired,
  items: PropTypes.arrayOf(PropTypes.string).isRequired,
  value: PropTypes.string.isRequired,
  onChange: PropTypes.func.isRequired,
};

export default Select;
