import React from 'react';
import PropTypes from 'prop-types';
import Dropdown from 'react-bootstrap/Dropdown';


const Select = (props) => {
  const { label, items, value, onChange } = props;
  return (
    <div className="d-flex">
      <div style={{ margin: 'auto 0' }}>
        {label}
      </div>
      <Dropdown className="m-1 w-25">
        <Dropdown.Toggle
          className="w-50"
          variant="outline-primary"
        >
          {value}
        </Dropdown.Toggle>
        <Dropdown.Menu className="w-25">
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
    </div>
  );
};

Select.propTypes = {
  label: PropTypes.string.isRequired,
  items: PropTypes.arrayOf(PropTypes.string).isRequired,
  value: PropTypes.string.isRequired,
  onChange: PropTypes.func.isRequired,
};

export default Select;
