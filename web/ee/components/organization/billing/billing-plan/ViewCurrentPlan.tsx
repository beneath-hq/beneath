import { Button, Grid, makeStyles, Paper, Typography } from "@material-ui/core";
import { FC } from "react";
import Moment from "react-moment";

import VSpace from "components/VSpace";
import { BillingInfo_billingInfo } from "ee/apollo/types/BillingInfo";
import { toURLName } from "lib/names";
import { OrganizationByName_organizationByName_PrivateOrganization } from "apollo/types/OrganizationByName";

const useStyles = makeStyles((theme) => ({
  paper: {
    padding: theme.spacing(3),
  },
  textData: {
    fontWeight: "bold"
  },
}));

interface Props {
  billingInfo: BillingInfo_billingInfo;
  organization: OrganizationByName_organizationByName_PrivateOrganization;
}

const ViewCurrentPlan: FC<Props> = ({billingInfo, organization}) => {
  const classes = useStyles();

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
        <VSpace units={3} />
        {billingInfo.billingPlan.default && (
          <Button
            variant="contained"
            color="primary"
            href={`/organization/-/billing/checkout?organization_name=${toURLName(organization.name)}`}
          >
            Upgrade plan
          </Button>
        )}
        {!billingInfo.billingPlan.default && (
          <Button
            variant="contained"
            href={`/organization/-/billing/checkout?organization_name=${toURLName(organization.name)}`}
          >
            Change plan
          </Button>
        )}
      </Paper>
    </>
  );
};

export default ViewCurrentPlan;