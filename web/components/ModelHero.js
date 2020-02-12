import PropTypes from "prop-types";
import { makeStyles } from "@material-ui/core/styles";

import Grid from "@material-ui/core/Grid";
import Typography from "@material-ui/core/Typography";

const useStyles = makeStyles((theme) => ({
  container: {
    marginTop: theme.spacing(4),
    marginBottom: theme.spacing(4),
  },
  site: {
    display: "block",
  },
}));

const ModelHero = ({ name, description }) => {
  const classes = useStyles();
  return (
    <Grid container wrap="nowrap" spacing={1} className={classes.container}>
      <Grid item>
        <Typography component="h1" variant="h1" gutterBottom>{name}</Typography>
        <Typography variant="body1">{description}</Typography>
      </Grid>
    </Grid>
  );
};

ModelHero.propTypes = {
  name: PropTypes.string,
  description: PropTypes.string,
};

export default ModelHero;
