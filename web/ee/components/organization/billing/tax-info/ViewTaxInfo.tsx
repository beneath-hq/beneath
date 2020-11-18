import { useQuery } from "@apollo/client";
import _ from "lodash";
import { Button, Grid, makeStyles, Paper, Table, TableBody, TableCell, TableHead, TableRow, Typography } from "@material-ui/core";
import React, { FC } from "react";

import { QUERY_BILLING_INFO } from "ee/apollo/queries/billingInfo";
import { BillingInfo, BillingInfoVariables } from "ee/apollo/types/BillingInfo";
import { OrganizationByName_organizationByName_PrivateOrganization } from "apollo/types/OrganizationByName";
import VSpace from "components/VSpace";

const useStyles = makeStyles((theme) => ({
  paperPadding: {
    padding: theme.spacing(3)
  },
  textData: {
    fontWeight: "bold",
  }
}));

export interface BillingInfoProps {
  organization: OrganizationByName_organizationByName_PrivateOrganization;
  editable?: boolean;
  editTaxInfo?: (value: boolean) => void;
}

const ViewTaxInfo: FC<BillingInfoProps> = ({ organization, editable, editTaxInfo }) => {
  const classes = useStyles();

  const { loading, error, data } = useQuery<BillingInfo, BillingInfoVariables>(QUERY_BILLING_INFO, {
    context: { ee: true },
    variables: {
      organizationID: organization.organizationID,
    },
  });

  if (error) {
    return <p>Error: {JSON.stringify(error)}</p>;
  }

  if (!data) {
    return <></>;
  }

  const rows = [
    {key: "Country", value: data.billingInfo.country},
    {key: "State", value: data.billingInfo.region},
    {key: "Company", value: data.billingInfo.companyName},
    {key: "Tax ID", value: data.billingInfo.taxNumber },
  ];

  return (
    <>
    <Grid container>
      <Grid item>
        <Paper className={classes.paperPadding} variant="outlined">
          {rows.map((row) => (
            <React.Fragment key={row.key}>
              <Grid container alignItems="center" spacing={1}>
                <Grid item>
                  <Typography>
                    {row.key}:
                  </Typography>
                </Grid>
                <Grid item>
                  <Typography className={classes.textData}>
                    {row.value}
                  </Typography>
                </Grid>
              </Grid>
              <VSpace units={1} />
            </React.Fragment>
          ))}
          {editable && editTaxInfo && (
            <>
              <VSpace units={3} />
              <Grid container justify="space-between">
                {/* <Grid item></Grid> */}
                <Grid item>
                  <Button onClick={() => editTaxInfo(true)} variant="contained">
                    Edit
                  </Button>
                </Grid>
              </Grid>
            </>
          )}
        </Paper>
      </Grid>
    </Grid>
    </>
  );
};

export default ViewTaxInfo;
