import React from 'react';
import PropTypes from 'prop-types';
import Dropdown from 'react-bootstrap/Dropdown';
import InputGroup from 'react-bootstrap/InputGroup';
import OverlayTrigger from 'react-bootstrap/OverlayTrigger';
import Tooltip from 'react-bootstrap/Tooltip';

const Select = (props) => {
  const {
    label,
    items,
    value,
    parentOnChange,
    tooltip,
  } = props;
  return (
    <InputGroup className="d-flex m-1 w-100">
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
                onClick={parentOnChange}
                value={item}
                key={item}
              >
                {item}
              </Dropdown.Item>
            ))
          }
        </Dropdown.Menu>
      </Dropdown>
      {tooltip && (
        <OverlayTrigger overlay={<Tooltip id="tooltip-disabled">{tooltip}</Tooltip>}>
          <span className="cc-tooltip ml-1 mr-1">
            <div className="cc-tooltip__icon">?</div>
          </span>
        </OverlayTrigger>
      )}
    </InputGroup>
  );
};

Select.propTypes = {
  label: PropTypes.string.isRequired,
  items: PropTypes.arrayOf(PropTypes.string).isRequired,
  value: PropTypes.string.isRequired,
  parentOnChange: PropTypes.func.isRequired,
  tooltip: PropTypes.string,
};

Select.defaultProps = {
  tooltip: '',
};

export default Select;
