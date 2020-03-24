import React, { useState } from 'react';
import PropTypes from 'prop-types';
import FormControl from 'react-bootstrap/FormControl';
import InputGroup from 'react-bootstrap/InputGroup';
import OverlayTrigger from 'react-bootstrap/OverlayTrigger';
import Tooltip from 'react-bootstrap/Tooltip';


const HexKey = (props) => {
  const {
    label,
    value,
    values,
    parentOnChange,
    tooltip,
  } = props;
  const [inputValid, setInputValid] = useState(true);
  return (
    <div>
      <InputGroup className={`cc-ip-input m-1 w-100 ${inputValid ? '' : 'cc-ip-input--invalid'}`}>
        <InputGroup.Prepend>
          <InputGroup.Text className="input-text">{label}</InputGroup.Text>
        </InputGroup.Prepend>
        <FormControl
          value={value}
          onChange={(e) => {
            // Only check for valid input on change
            if (values !== [] && values.includes(value.length)) {
              setInputValid(true);
            }
            parentOnChange(e);
          }}
          onBlur={(e) => {
            // Only check for invalid input on defocus
            if (values !== [] && !values.includes(value.length)) {
              setInputValid(false);
            }
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
      {!inputValid && (
        <p className="cc-ip-input__err-text ml-1">
          Please enter a hex key with one of the following lengths:
          {` ${values.join(', ')}`}
        </p>
      )}
    </div>
  );
};

HexKey.propTypes = {
  label: PropTypes.string.isRequired,
  value: PropTypes.string,
  values: PropTypes.arrayOf(PropTypes.number),
  parentOnChange: PropTypes.func.isRequired,
  tooltip: PropTypes.string,
};

HexKey.defaultProps = {
  value: '',
  values: [],
  tooltip: '',
};

export default HexKey;
