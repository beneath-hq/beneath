import React, { FC } from "react";
import { Button, Grid, makeStyles, Theme, Typography } from "@material-ui/core";

import { Tile, TileProps } from "./Tile";
import { NakedLink } from "components/Link";
import useMe from "hooks/useMe";
import { useQuery } from "@apollo/client";
import { BillingInfo, BillingInfoVariables } from "apollo/types/BillingInfo";
import { QUERY_BILLING_INFO } from "apollo/queries/billinginfo";

const useStyles = makeStyles((theme: Theme) => ({
  container: {
    padding: theme.spacing(2),
  },
  description: {
    marginTop: theme.spacing(1.5),
  },
  buttons: {
    marginTop: theme.spacing(1.5)
  }
}));

export const UpgradeTile: FC<TileProps> = ({ ...tileProps }) => {
  const classes = useStyles();
  const me = useMe();

  if (!me || !me.personalUserID) {
    return <></>;
  }

  const { loading, error, data } = useQuery<BillingInfo, BillingInfoVariables>(QUERY_BILLING_INFO, {
    variables: {
      organizationID: me.organizationID,
    },
  });

  if (error || !data) {
    return <p>Error: {JSON.stringify(error)}</p>;
  }

  return (
    <>
      {data.billingInfo.billingPlan.default && (
        <Tile {...tileProps}>
          <Grid className={classes.container} container justify="center" alignContent="center" alignItems="center" direction="column">
            <Grid item xs>
              <Typography variant="h3" align="center">
                Upgrade to Pro
              </Typography>
              <Typography  gutterBottom align="center" className={classes.description}>
                Get higher quotas and premium support.
              </Typography>
              </Grid>
            <Grid item>
              <Grid container spacing={1} className={classes.buttons}>
                <Grid item>
                  <Button variant="contained" href={"https://about.beneath.dev/contact"}>Contact us</Button>
                </Grid>
                <Grid item>
                  <Button
                    color="primary"
                    variant="contained"
                    component={NakedLink}
                    href={`/organization?organization_name=${me.name}&tab=billing`}
                    >
                    Upgrade
                  </Button>
                </Grid>
              </Grid>
            </Grid>
          </Grid>
        </Tile>
      )}
    </>
  );
};

export default UpgradeTile;
