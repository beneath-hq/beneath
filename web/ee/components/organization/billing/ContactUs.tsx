import { Button, Grid, makeStyles, Paper, Typography } from "@material-ui/core";
import { FC } from "react";

const useStyles = makeStyles((theme) => ({
  paperPadding: {
    padding: theme.spacing(3),
  },
  headline: {
    marginBottom: theme.spacing(2)
  },
  button: {
    marginTop: theme.spacing(3)
  }
}));

const ContactUs: FC = () => {
  const classes = useStyles();

  return (
    <>
      {/* <Grid item> */}
      <Paper variant="outlined" className={classes.paperPadding}>
        <Typography>
          We're here to talk about <strong>custom Enterprise plans</strong>, <strong>discounts for public data</strong>, or anything else.
        </Typography>
        <Button
          variant="contained"
          className={classes.button}
          href="https://docs.google.com/forms/d/e/1FAIpQLSdsO3kcT3yk0Cgc4MzkPR_d16jZiYQd7L0M3ZxGwdOYycGhIg/viewform?usp=sf_link"
          >
          Contact us
        </Button>
      </Paper>
      {/* </Grid> */}
    </>
  );
};

export default ContactUs;


