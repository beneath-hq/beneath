import { Button, Grid, makeStyles, Paper, Typography } from "@material-ui/core";
import { BillingInfo_billingInfo } from "ee/apollo/types/BillingInfo";
import { PROFESSIONAL_BOOST_PLAN } from "ee/lib/billing";
import { FC } from "react";

const useStyles = makeStyles((theme) => ({
  paper: {
    padding: theme.spacing(3),
  },
  headline: {
    marginBottom: theme.spacing(2)
  },
  button: {
    marginTop: theme.spacing(3)
  }
}));

interface Props {
  billingInfo: BillingInfo_billingInfo;
}

const ContactUs: FC<Props> = ({billingInfo}) => {
  const classes = useStyles();

  return (
    <>
      <Paper variant="outlined" className={classes.paper}>
        <Typography>
          We're here to talk about <strong>custom Enterprise plans</strong>, <strong>discounts for public data</strong>, or anything else.
        </Typography>
        {billingInfo.billingPlan.description !== PROFESSIONAL_BOOST_PLAN && (
          <Button
            variant="contained"
            className={classes.button}
            href="https://docs.google.com/forms/d/e/1FAIpQLSdsO3kcT3yk0Cgc4MzkPR_d16jZiYQd7L0M3ZxGwdOYycGhIg/viewform?usp=sf_link"
            >
            Contact us
          </Button>
        )}
        {billingInfo.billingPlan.description === PROFESSIONAL_BOOST_PLAN && (
          <Button
            variant="contained"
            color="primary"
            className={classes.button}
            href="https://docs.google.com/forms/d/e/1FAIpQLSdsO3kcT3yk0Cgc4MzkPR_d16jZiYQd7L0M3ZxGwdOYycGhIg/viewform?usp=sf_link"
            >
            Go enterprise
          </Button>
        )}
      </Paper>
    </>
  );
};

export default ContactUs;


