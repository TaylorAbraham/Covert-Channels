import React from 'react';
import PropTypes from 'prop-types';

const Checkbox = (props) => {
  const { label, value, onChange } = props;
  return (
    <label className="check-container">
      <div className="check-label">{label}</div>
      <input
        type="checkbox"
        checked={value}
        onChange={onChange}
      />
      <span className="checkmark" />
    </label>
  );
};

Checkbox.propTypes = {
  label: PropTypes.string.isRequired,
  value: PropTypes.bool,
  onChange: PropTypes.func.isRequired,
};

Checkbox.defaultProps = {
  value: false,
};

export default Checkbox;
