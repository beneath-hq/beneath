import { Button, Grid, makeStyles, Typography } from "@material-ui/core";
import { FC } from "react";

import { OrganizationByName_organizationByName_PrivateOrganization } from "apollo/types/OrganizationByName";
import { BillingInfo_billingInfo } from "ee/apollo/types/BillingInfo";
import { toURLName } from "lib/names";

const useStyles = makeStyles((theme) => ({
  button: {
    marginTop: theme.spacing(3)
  }
}));

interface Props {
  billingInfo: BillingInfo_billingInfo;
  organization: OrganizationByName_organizationByName_PrivateOrganization;
}

const Actions: FC<Props> = ({billingInfo, organization}) => {
  const classes = useStyles();

  return (
    <>
      <Typography variant="h2" gutterBottom>
        Actions
      </Typography>
      <Grid container spacing={3}>
        <Grid item xs={6}>
          {billingInfo.billingPlan.default && (
            <>
              <Typography>
                Level up your data pipelines and get <strong>expanded quotas</strong> for your data projects.
              </Typography>
              <Button
                variant="contained"
                className={classes.button}
                color="primary"
                href={`/organization/-/billing/checkout?organization_name=${toURLName(organization.name)}`}
              >
                Upgrade plan
              </Button>
            </>
          )}
          {!billingInfo.billingPlan.default && (
            <>
              <Typography>
                Switch your plan to get the best fit for you.
              </Typography>
              <Button
                variant="contained"
                className={classes.button}
                href={`/organization/-/billing/checkout?organization_name=${toURLName(organization.name)}`}
              >
                Change plan
              </Button>
            </>
          )}
        </Grid>
        <Grid item xs={6}>
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
        </Grid>
      </Grid>
    </>
  );
};

export default Actions;


