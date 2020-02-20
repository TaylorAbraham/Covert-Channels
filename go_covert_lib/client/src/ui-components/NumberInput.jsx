import React from 'react';
import PropTypes from 'prop-types';
import FormControl from 'react-bootstrap/FormControl';
import InputGroup from 'react-bootstrap/InputGroup';
import OverlayTrigger from 'react-bootstrap/OverlayTrigger';
import Tooltip from 'react-bootstrap/Tooltip';

const NumberInput = (props) => {
  const {
    label,
    value,
    min,
    max,
    parentOnChange,
    tooltip,
  } = props;
  return (
    <InputGroup className="m-1 w-100">
      <InputGroup.Prepend>
        <InputGroup.Text className="input-text">{label}</InputGroup.Text>
      </InputGroup.Prepend>
      <FormControl
        value={value}
        onChange={(e) => {
          if (e.target.value > max) {
            e.target.value = max;
          } else if (e.target.value < min) {
            e.target.value = min;
          }
          parentOnChange(e);
        }}
      />
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

NumberInput.propTypes = {
  label: PropTypes.string.isRequired,
  value: PropTypes.number,
  min: PropTypes.number,
  max: PropTypes.number,
  parentOnChange: PropTypes.func.isRequired,
  tooltip: PropTypes.string,
};

NumberInput.defaultProps = {
  value: 0,
  min: Number.MIN_SAFE_INTEGER,
  max: Number.MAX_SAFE_INTEGER,
  tooltip: '',
};

export default NumberInput;
