import PropTypes from "prop-types";
import pure from "recompose/pure";
import { makeStyles } from "@material-ui/core/styles";

const useStyles = makeStyles((theme) => ({
  space: ({ units }) => ({
    display: "block",
    paddingTop: theme.spacing(units),
  }),
}));

let VSpace = (props) => {
  let units = props.units || 1;
  const classes = useStyles({ units });
  return <div className={classes.space} />;
};

VSpace.displayName = "VSpace";
VSpace = pure(VSpace);
VSpace.muiName = "VSpace";

VSpace.propTypes = {
  units: PropTypes.number,
  sidebar: PropTypes.object,
};

export default VSpace;
