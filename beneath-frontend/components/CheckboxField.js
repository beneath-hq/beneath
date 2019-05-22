import PropTypes from "prop-types";
import { makeStyles } from "@material-ui/core/styles";

import Checkbox from "@material-ui/core/Checkbox";
import FormControlLabel from "@material-ui/core/FormControlLabel";

const useStyles = makeStyles((theme) => ({
}));

const CheckboxField = ({ label, checked, onChange }) => {
  const classes = useStyles();
  return (
    <FormControlLabel label={label} control={<Checkbox checked={checked} onChange={onChange} />}
    />
  );
};

CheckboxField.propTypes = {
  label: PropTypes.string,
  checked: PropTypes.bool,
  onChange: PropTypes.func,
};

export default CheckboxField;
