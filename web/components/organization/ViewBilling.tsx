import _ from "lodash";
import React, { FC } from "react";

import { Grid, Link, makeStyles, Paper, Typography } from "@material-ui/core";

import { OrganizationByName_organizationByName_PrivateOrganization } from "../../apollo/types/OrganizationByName";
import ViewBillingInfo from "./billing/ViewBillingInfo";
import ViewBillingMethods from "./billing/ViewBillingMethods";

const useStyles = makeStyles((theme) => ({
  banner: {
    padding: theme.spacing(2),
  },
}));

export interface ViewBillingProps {
  organization: OrganizationByName_organizationByName_PrivateOrganization;
}

const ViewBilling: FC<ViewBillingProps> = ({ organization }) => {
  const classes = useStyles();

  const specialCase =
    organization.personalUser && organization.personalUser.billingOrganizationID !== organization.organizationID;

  return (
    <React.Fragment>
      {specialCase && (
        <Paper elevation={1} square>
          <Typography className={classes.banner}>
            You are part of an organization that handles your billing. The billing information on this page
             manages any of your resources that you did not transfer to your organization.
          </Typography>
        </Paper>
      )}

      <Paper elevation={1} square>
        <Typography className={classes.banner}>
          You can find detailed information about our billing plans {" "}
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
