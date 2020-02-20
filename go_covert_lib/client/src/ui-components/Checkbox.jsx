import React from 'react';
import PropTypes from 'prop-types';
import OverlayTrigger from 'react-bootstrap/OverlayTrigger';
import Tooltip from 'react-bootstrap/Tooltip';
import InputGroup from 'react-bootstrap/InputGroup';

const Checkbox = (props) => {
  const {
    label,
    value,
    onChange,
    tooltip,
  } = props;
  return (
    <InputGroup className="ml-1">
      <label className="check-container">
        <div className="check-label">{label}</div>
        <input
          type="checkbox"
          checked={value}
          onChange={onChange}
        />
        <span className="checkmark" />
      </label>
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

Checkbox.propTypes = {
  label: PropTypes.string.isRequired,
  value: PropTypes.bool,
  onChange: PropTypes.func.isRequired,
  tooltip: PropTypes.string,
};

Checkbox.defaultProps = {
  value: false,
  tooltip: '',
};

export default Checkbox;
