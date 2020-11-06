import _ from "lodash";
import React, { FC } from "react";

import { Grid, Link, makeStyles, Paper, Typography } from "@material-ui/core";

import { OrganizationByName_organizationByName_PrivateOrganization } from "../../../apollo/types/OrganizationByName";
import ViewBillingInfo from "./billing/ViewBillingInfo";
import ViewBillingMethods from "./billing/ViewBillingMethods";

const useStyles = makeStyles((theme) => ({
  banner: {
    padding: theme.spacing(2),
    marginBottom: theme.spacing(2),
  },
}));

export interface ViewBillingProps {
  organization: OrganizationByName_organizationByName_PrivateOrganization;
}

const ViewBilling: FC<ViewBillingProps> = ({ organization }) => {
  const classes = useStyles();

  // alert when you're viewing billing for your personal organization but your main billing is handled by another org
  const specialCase =
    organization.personalUser && organization.personalUser.billingOrganizationID !== organization.organizationID;

  return (
    <React.Fragment>
      {specialCase && (
        <Paper elevation={1} square>
          <Typography className={classes.banner}>
            Note that you are a member of the {organization.personalUser?.billingOrganization.displayName} organization,
            and so {organization.personalUser?.billingOrganization.displayName} is billed for your Beneath usage.
            However, you are individually billed for the activity of any services that you have not transferred to{" "}
            {organization.personalUser?.billingOrganization.displayName}.
            <br />
            <br /> If you have no active services managed by your user, then make sure to cancel any active billing plan
            on this page. Then you won't incur any unneccessary charges.
            <br />
            <br />
            If you have active services that you would like {
              organization.personalUser?.billingOrganization.displayName
            }{" "}
            to pay for, you need to transfer those services from your personal user to the{" "}
            {organization.personalUser?.billingOrganization.displayName} organization. See the docs{" "}
            <Link href="https://about.beneath.dev/docs/core-resources/services">here</Link>.
            <br />
            <br /> If you have active services that you would like to pay for yourself, and not assign to your
            organization, then you must ensure the billing information on this page covers you. If the services' usage
            exceeds the quotas of the Free tier, then you should upgrade your personal billing plan on this page.
          </Typography>
        </Paper>
      )}

      <Paper elevation={1} square>
        <Typography className={classes.banner}>
          You can find detailed information about our billing plans{" "}
          <Link href="https://about.beneath.dev/enterprise">here</Link>.
        </Typography>
      </Paper>

      <Grid container direction="column">
        <Grid item>
          <ViewBillingInfo organization={organization} />
        </Grid>
        <Grid item>
          <ViewBillingMethods organization={organization} />
        </Grid>
      </Grid>
    </React.Fragment>
  );
};

export default ViewBilling;
