import React from 'react';
import PropTypes from 'prop-types';
import FormControl from 'react-bootstrap/FormControl';
import InputGroup from 'react-bootstrap/InputGroup';
import OverlayTrigger from 'react-bootstrap/OverlayTrigger';
import Tooltip from 'react-bootstrap/Tooltip';


const IPInput = (props) => {
  const {
    label,
    value,
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
        onChange={parentOnChange}
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

IPInput.propTypes = {
  label: PropTypes.string.isRequired,
  value: PropTypes.string,
  parentOnChange: PropTypes.func.isRequired,
  tooltip: PropTypes.string,
};

IPInput.defaultProps = {
  value: '',
  tooltip: '',
};

export default IPInput;
