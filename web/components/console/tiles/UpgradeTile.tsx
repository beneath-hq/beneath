import React, { FC } from "react";
import { Button, Grid, makeStyles, Theme, Typography } from "@material-ui/core";

import { Tile, TileProps } from "./Tile";
import { NakedLink } from "components/Link";
import useMe from "hooks/useMe";
import { useQuery } from "@apollo/client";
import { BillingInfo, BillingInfoVariables } from "ee/apollo/types/BillingInfo";
import { QUERY_BILLING_INFO } from "ee/apollo/queries/billingInfo";
import { IS_EE } from "lib/connection";
import { PaperGrid } from "components/Paper";

const useStyles = makeStyles((theme: Theme) => ({
  title: {
    marginBottom: "0.5rem",
  },
}));

export const UpgradeTile: FC<TileProps> = ({ ...tileProps }) => {
  const classes = useStyles();
  const me = useMe();

  const skip = !IS_EE || !me || me?.personalUser?.billingOrganizationID !== me.organizationID;
  const { data } = useQuery<BillingInfo, BillingInfoVariables>(QUERY_BILLING_INFO, {
    context: { ee: true },
    variables: {
      organizationID: me?.personalUser?.billingOrganizationID || "",
    },
    skip: skip,
  });

  if (skip || !data?.billingInfo.billingPlan.default) {
    return <></>;
  }

  return (
    <Tile {...tileProps} nopaper>
      <PaperGrid variant="outlined" container direction="column" alignContent="stretch">
        <Grid item>
          <Typography variant="h3" align="center" className={classes.title}>
            Upgrade to Pro
          </Typography>
          <Typography variant="body2" align="center" gutterBottom>
            Get higher quotas and premium support
          </Typography>
        </Grid>
        <Grid item>
          <Grid container spacing={1}>
            <Grid item xs={6}>
              <Button fullWidth variant="contained" href="https://about.beneath.dev/contact" target="_blank">
                Contact&nbsp;us
              </Button>
            </Grid>
            <Grid item xs={6}>
              <Button
                fullWidth
                color="primary"
                variant="contained"
                component={NakedLink}
                href={`/organization?organization_name=${me?.name}&tab=billing`}
                as={`${me?.name}/-/billing`}
              >
                Upgrade
              </Button>
            </Grid>
          </Grid>
        </Grid>
      </PaperGrid>
    </Tile>
  );
};

export default UpgradeTile;
