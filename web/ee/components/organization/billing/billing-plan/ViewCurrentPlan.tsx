import { Button, Grid, makeStyles, Paper, Typography, useMediaQuery, useTheme } from "@material-ui/core";
import VSpace from "components/VSpace";
import { BillingInfo_billingInfo } from "ee/apollo/types/BillingInfo";
import { PROFESSIONAL_BOOST_PLAN } from "ee/lib/billing";
import { FC } from "react";
import Moment from "react-moment";

const useStyles = makeStyles((theme) => ({
  paper: {
    padding: theme.spacing(3),
  },
  textData: {
    fontWeight: "bold"
  },
  buttons: {
    [theme.breakpoints.up("md")]: {
      marginTop: theme.spacing(1),
    },
  }
}));

interface Props {
  billingInfo: BillingInfo_billingInfo;
  cancelPlan: (value: boolean) => void;
  changePlan: (value: boolean) => void;
}

const ViewCurrentPlan: FC<Props> = ({billingInfo, cancelPlan, changePlan}) => {
  const classes = useStyles();
  const theme = useTheme();
  const isXs = useMediaQuery(theme.breakpoints.down("xs"));

  return (
    <>
      <Paper className={classes.paper} variant="outlined">
        <Grid container spacing={1}  alignItems="center">
          <Grid item>
            <Typography>
              Plan name:
            </Typography>
          </Grid>
          <Grid item>
            <Typography className={classes.textData}>
              {billingInfo.billingPlan.description}
            </Typography>
          </Grid>
        </Grid>
        <VSpace units={1} />
        <Grid container spacing={1} >
          <Grid item>
            <Typography>
              Current billing period:
            </Typography>
          </Grid>
          <Grid item>
            <Typography className={classes.textData}>
              <Moment format="MMMM Do" subtract={{ days: 31 }}>{billingInfo.nextBillingTime}</Moment> to <Moment format="MMMM Do">{billingInfo.nextBillingTime}</Moment>
            </Typography>
          </Grid>
        </Grid>
        <VSpace units={3} />
        <Grid container spacing={2} className={classes.buttons} alignItems="center">
          {!billingInfo.billingPlan.default && (
            <Grid item>
              <Button size={isXs ? 'small' : 'medium'} onClick={() => cancelPlan(true)}>Cancel plan</Button>
            </Grid>
          )}
          <Grid item>
            {billingInfo.billingPlan.description !== PROFESSIONAL_BOOST_PLAN && (
              <Button size={(isXs && !billingInfo.billingPlan.default) ? 'small' : 'medium'} variant="contained" color="primary" onClick={() => changePlan(true)}>
                Upgrade plan
              </Button>
            )}
            {billingInfo.billingPlan.description === PROFESSIONAL_BOOST_PLAN && (
              <Button size={isXs ? 'small' : 'medium'} variant="contained" onClick={() => changePlan(true)}>
                Change plan
              </Button>
            )}
          </Grid>
        </Grid>
      </Paper>
    </>
  );
};

export default ViewCurrentPlan;