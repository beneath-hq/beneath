import { Button, Grid, makeStyles, Paper, Typography } from "@material-ui/core";
import VSpace from "components/VSpace";
import { BillingInfo_billingInfo } from "ee/apollo/types/BillingInfo";
import { FC } from "react";
import Moment from "react-moment";

const useStyles = makeStyles((theme) => ({
  paperPadding: {
    padding: theme.spacing(3)
  },
  textData: {
    fontWeight: "bold"
  },
}));

interface Props {
  billingInfo: BillingInfo_billingInfo;
  cancelPlan: (value: boolean) => void;
  changePlan: (value: boolean) => void;
}

const ViewCurrentPlan: FC<Props> = ({billingInfo, cancelPlan, changePlan}) => {
  const classes = useStyles();

  return (
    <>
      <Paper className={classes.paperPadding} variant="outlined">
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
        <Grid container spacing={2}>
          {!billingInfo.billingPlan.default && (
            <Grid item>
              <Button onClick={() => cancelPlan(true)}>Cancel plan</Button>
            </Grid>
          )}
          <Grid item>
            <Button variant="contained" color="primary" onClick={() => changePlan(true)}>
              Upgrade plan
            </Button>
          </Grid>
        </Grid>
      </Paper>
    </>
  );
};

export default ViewCurrentPlan;