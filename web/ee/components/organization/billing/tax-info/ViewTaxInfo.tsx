import { useQuery } from "@apollo/client";
import _ from "lodash";
import { Button, Grid, makeStyles, Paper, Typography } from "@material-ui/core";
import React, { FC } from "react";

import { QUERY_BILLING_INFO } from "ee/apollo/queries/billingInfo";
import { BillingInfo, BillingInfoVariables } from "ee/apollo/types/BillingInfo";
import { OrganizationByName_organizationByName_PrivateOrganization } from "apollo/types/OrganizationByName";
import VSpace from "components/VSpace";
import ContentContainer from "components/ContentContainer";

const useStyles = makeStyles((theme) => ({
  paper: {
    padding: theme.spacing(3),
    height: "100%",
  },
  paperTitle: {
    marginBottom: theme.spacing(1),
  },
  textData: {
    fontWeight: "bold",
  },
  button: {
    marginTop: theme.spacing(3),
  },
}));

export interface BillingInfoProps {
  organization: OrganizationByName_organizationByName_PrivateOrganization;
  onEdit?: () => void;
  littleHeader?: boolean;
}

const ViewTaxInfo: FC<BillingInfoProps> = ({ organization, onEdit, littleHeader }) => {
  const classes = useStyles();

  const { loading, error, data } = useQuery<BillingInfo, BillingInfoVariables>(QUERY_BILLING_INFO, {
    context: { ee: true },
    variables: {
      organizationID: organization.organizationID,
    },
  });

  if (!data || error) return null;
  const billingInfo = data.billingInfo;

  // construct the view
  let rows = [{ key: "Country", value: billingInfo.country }];
  if (billingInfo.country === "United States of America") {
    rows = rows.concat({ key: "State", value: billingInfo.region as string });
  }
  if (billingInfo.taxNumber) {
    rows = rows.concat(
      { key: "Tax entity", value: "Company" },
      { key: "Company", value: billingInfo.companyName as string },
      { key: "Tax ID", value: billingInfo.taxNumber as string }
    );
  } else {
    rows = rows.concat({ key: "Tax entity", value: "Individual" });
  }

  let empty: boolean | undefined;
  if (!billingInfo.country && !billingInfo.companyName && !billingInfo.taxNumber) {
    empty = true;
  }

  return (
    <>
      <Paper variant="outlined" className={classes.paper}>
        <Typography variant={littleHeader ? "h2" : "h1"} className={classes.paperTitle}>
          Tax info
        </Typography>
        <Typography variant="body2" color="textSecondary">
          Information used to compute tax for customers in certain countries
        </Typography>
        <VSpace units={3} />
        {empty && onEdit && (
          // TODO: replace ContentContainer with just a nicely formatted callToAction for TitledPaper components
          <ContentContainer
            callToAction={{
              message: `You have not provided any tax information`,
              buttons: [{ label: "Edit", onClick: () => onEdit() }],
            }}
          />
        )}
        {!empty && (
          <>
            {rows.map((row) => (
              <React.Fragment key={row.key}>
                <Grid container alignItems="center" spacing={1}>
                  <Grid item>
                    <Typography>{row.key}:</Typography>
                  </Grid>
                  <Grid item>
                    <Typography className={classes.textData}>{row.value}</Typography>
                  </Grid>
                </Grid>
              </React.Fragment>
            ))}
            {onEdit && (
              <>
                <Button onClick={() => onEdit()} variant="contained" className={classes.button}>
                  Edit
                </Button>
              </>
            )}
          </>
        )}
      </Paper>
    </>
  );
};

export default ViewTaxInfo;
