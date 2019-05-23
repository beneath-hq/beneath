import PropTypes from "prop-types";
import { makeStyles } from "@material-ui/core/styles";

import Checkbox from "@material-ui/core/Checkbox";
import FormControl from "@material-ui/core/FormControl";
import FormControlLabel from "@material-ui/core/FormControlLabel";
import FormHelperText from "@material-ui/core/FormHelperText";

const useStyles = makeStyles((theme) => ({
}));

const CheckboxField = ({ label, checked, onChange, helperText, ...other }) => {
  const classes = useStyles();
  return (
    <FormControl {...other}>
      <FormControlLabel label={label} control={<Checkbox checked={checked} onChange={onChange} />} />
      {helperText && <FormHelperText>{helperText}</FormHelperText>}
    </FormControl>
  );
};

CheckboxField.propTypes = {
  label: PropTypes.string,
  checked: PropTypes.bool,
  onChange: PropTypes.func,
};

export default CheckboxField;
