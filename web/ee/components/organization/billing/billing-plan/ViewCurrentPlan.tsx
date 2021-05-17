import { Button, Grid, makeStyles, Typography } from "@material-ui/core";
import { FC } from "react";
import Moment from "react-moment";

import { OrganizationByName_organizationByName_PrivateOrganization } from "apollo/types/OrganizationByName";
import { NakedLink } from "components/Link";
import VSpace from "components/VSpace";
import { BillingInfo_billingInfo } from "ee/apollo/types/BillingInfo";
import { CONTACT_LINK } from "ee/lib/billing";
import { toURLName } from "lib/names";

const useStyles = makeStyles((theme) => ({
  sectionHeader: {
    marginTop: theme.spacing(3),
    marginBottom: theme.spacing(2),
  },
  textData: {
    fontWeight: "bold",
  },
  buttons: {
    marginTop: theme.spacing(3),
    marginBottom: theme.spacing(1),
  },
}));

interface Props {
  organization: OrganizationByName_organizationByName_PrivateOrganization;
  billingInfo: BillingInfo_billingInfo;
}

const ViewCurrentPlan: FC<Props> = ({ organization, billingInfo }) => {
  const classes = useStyles();

  return (
    <>
      <Typography variant="h2" className={classes.sectionHeader}>
        Your current plan
      </Typography>
      <Grid container spacing={1} alignItems="center">
        <Grid item>
          <Typography>Plan name:</Typography>
        </Grid>
        <Grid item>
          <Typography className={classes.textData}>{billingInfo.billingPlan.name}</Typography>
        </Grid>
      </Grid>
      <VSpace units={1} />
      <Grid container spacing={1}>
        <Grid item>
          <Typography>Current billing period:</Typography>
        </Grid>
        <Grid item>
          <Typography className={classes.textData}>
            <Moment format="MMMM Do" subtract={{ days: 31 }}>
              {billingInfo.nextBillingTime}
            </Moment>{" "}
            to <Moment format="MMMM Do">{billingInfo.nextBillingTime}</Moment>
          </Typography>
        </Grid>
      </Grid>
      <Grid container spacing={2} className={classes.buttons}>
        <Grid item>
          {billingInfo.billingPlan.default && (
            <>
              <Button
                variant="contained"
                color="primary"
                component={NakedLink}
                href={`/-/billing/checkout?organization_name=${toURLName(organization.name)}`}
                as={`${toURLName(organization.name)}/-/billing/checkout`}
              >
                Upgrade plan
              </Button>
            </>
          )}
          {!billingInfo.billingPlan.default && (
            <>
              <Button
                variant="contained"
                component={NakedLink}
                href={`/-/billing/checkout?organization_name=${toURLName(organization.name)}`}
                as={`${toURLName(organization.name)}/-/billing/checkout`}
              >
                Change plan
              </Button>
            </>
          )}
        </Grid>
        <Grid item>
          <Button variant="contained" href={CONTACT_LINK} target="_blank">
            Contact sales
          </Button>
        </Grid>
      </Grid>
    </>
  );
};

export default ViewCurrentPlan;
