import { Grid, makeStyles, Typography } from "@material-ui/core";
import { FC } from "react";
import Moment from "react-moment";

import VSpace from "components/VSpace";
import { BillingInfo_billingInfo } from "ee/apollo/types/BillingInfo";

const useStyles = makeStyles((theme) => ({
  textData: {
    fontWeight: "bold"
  },
}));

interface Props {
  billingInfo: BillingInfo_billingInfo;
}

const ViewCurrentPlan: FC<Props> = ({billingInfo}) => {
  const classes = useStyles();

  return (
    <>
      <Typography variant="h2" gutterBottom>
        Your current plan
      </Typography>
      <Grid container spacing={1}  alignItems="center">
        <Grid item>
          <Typography>
            Plan name:
          </Typography>
        </Grid>
        <Grid item>
          <Typography className={classes.textData}>
            {billingInfo.billingPlan.name}
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
    </>
  );
};

export default ViewCurrentPlan;