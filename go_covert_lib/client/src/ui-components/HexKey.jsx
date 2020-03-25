import React, { useState, useEffect } from 'react';
import PropTypes from 'prop-types';
import FormControl from 'react-bootstrap/FormControl';
import InputGroup from 'react-bootstrap/InputGroup';
import OverlayTrigger from 'react-bootstrap/OverlayTrigger';
import Tooltip from 'react-bootstrap/Tooltip';


const HexKey = (props) => {
  const {
    label,
    value,
    acceptedLengths,
    parentOnChange,
    tooltip,
  } = props;
  const [inputValid, setInputValid] = useState(true);
  const [displayedVal, setDisplayedVal] = useState('');
  const regex = /^([a-fA-F0-9][a-fA-F0-9])*$/;
  useEffect(() => {
    const sval = atob(value);
    let newDisplayedVal = '';
    // Inspired by https://stackoverflow.com/questions/14603205/how-to-convert-hex-string-into-a-bytes-array-and-a-bytes-array-in-the-hex-string
    for (let i = 0; i < sval.length; i++) {
      const c = sval.charCodeAt(i).toString(16);
      newDisplayedVal += c.length === 1 ? (`0${c}`) : (c);
    }
    setDisplayedVal(newDisplayedVal);
  }, []);
  return (
    <div>
      <InputGroup className={`cc-ip-input m-1 w-100 ${inputValid ? '' : 'cc-ip-input--invalid'}`}>
        <InputGroup.Prepend>
          <InputGroup.Text className="input-text">{label}</InputGroup.Text>
        </InputGroup.Prepend>
        <FormControl
          value={displayedVal}
          onChange={(e) => {
            const key = [];
            let str = e.target.value;
            // Inspired by https://stackoverflow.com/questions/14603205/how-to-convert-hex-string-into-a-bytes-array-and-a-bytes-array-in-the-hex-string
            if (str.match(regex) !== null) {
              while (str.length > 0) {
                key.push(parseInt(str.substring(0, 2), 16));
                str = str.substring(2);
              }
              parentOnChange({
                ...e,
                target: {
                  ...e.target,
                  value: key,
                },
              });
            }
            // Only check for valid input on change
            if (acceptedLengths !== [] && acceptedLengths.includes(key.length)) {
              setInputValid(true);
            }
            setDisplayedVal(e.target.value);
          }}
          onBlur={() => {
            // Only check for invalid input on defocus
            if ((acceptedLengths !== [] && !acceptedLengths.includes(value.length))
                || !displayedVal.match(regex)) {
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
          {` ${acceptedLengths.join(', ')}`}
        </p>
      )}
    </div>
  );
};

HexKey.propTypes = {
  label: PropTypes.string.isRequired,
  value: PropTypes.oneOfType([PropTypes.string, PropTypes.array]),
  acceptedLengths: PropTypes.arrayOf(PropTypes.number),
  parentOnChange: PropTypes.func.isRequired,
  tooltip: PropTypes.string,
};

HexKey.defaultProps = {
  value: '',
  acceptedLengths: [],
  tooltip: '',
};

export default HexKey;
