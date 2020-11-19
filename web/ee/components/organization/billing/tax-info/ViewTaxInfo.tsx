import { useQuery } from "@apollo/client";
import _ from "lodash";
import { Button, Grid, makeStyles, Paper, Table, TableBody, TableCell, TableHead, TableRow, Typography } from "@material-ui/core";
import React, { FC } from "react";

import { QUERY_BILLING_INFO } from "ee/apollo/queries/billingInfo";
import { BillingInfo, BillingInfoVariables } from "ee/apollo/types/BillingInfo";
import { OrganizationByName_organizationByName_PrivateOrganization } from "apollo/types/OrganizationByName";
import VSpace from "components/VSpace";
import ContentContainer, { CallToAction } from "components/ContentContainer";

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

  // construct the table
  let rows = [{key: "Country", value: data.billingInfo.country}];
  if (data.billingInfo.country === "United States of America") {
    rows = rows.concat({key: "State", value: data.billingInfo.region as string});
  }
  if (data.billingInfo.taxNumber) {
    rows = rows.concat(
      {key: "Entity", value: "Company"},
      {key: "Company", value: data.billingInfo.companyName as string},
      {key: "Tax ID", value: data.billingInfo.taxNumber as string},
    );
  } else {
    rows = rows.concat({key: "Entity", value: "Individual"});
  }

  if (editable && editTaxInfo && !data.billingInfo.country && !data.billingInfo.companyName && !data.billingInfo.taxNumber) {
    const cta: CallToAction = {
      message: `You have not provided any tax information`,
      buttons: [{ label: "Add tax info", onClick: () => editTaxInfo(true) }]
    };
    return (
      <ContentContainer callToAction={cta} />
    );
  }

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
              <Button onClick={() => editTaxInfo(true)} variant="contained">
                Edit
              </Button>
            </>
          )}
        </Paper>
      </Grid>
    </Grid>
    </>
  );
};

export default ViewTaxInfo;
