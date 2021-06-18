import { useMutation, useQuery } from "@apollo/client";
import _ from "lodash";
import {
  Button,
  Grid,
  makeStyles,
  Paper,
  Table as MuiTable,
  TableBody as MuiTableBody,
  TableCell as MuiTableCell,
  TableHead as MuiTableHead,
  TableRow as MuiTableRow,
  Typography,
} from "@material-ui/core";
import { MoreVert } from "@material-ui/icons";
import React, { FC } from "react";

import { OrganizationByName_organizationByName_PrivateOrganization } from "apollo/types/OrganizationByName";
import ContentContainer from "components/ContentContainer";
import DropdownButton from "components/DropdownButton";
import { QUERY_BILLING_METHODS } from "ee/apollo/queries/billingMethod";
import { BillingMethods, BillingMethodsVariables } from "ee/apollo/types/BillingMethods";
import { BillingInfo_billingInfo, BillingInfo_billingInfo_billingMethod } from "ee/apollo/types/BillingInfo";
import { UpdateBillingMethod, UpdateBillingMethodVariables } from "ee/apollo/types/UpdateBillingMethod";
import { UPDATE_BILLING_METHOD } from "ee/apollo/queries/billingInfo";
import { ANARCHISM_DRIVER, STRIPECARD_DRIVER, STRIPEWIRE_DRIVER } from "ee/lib/billing";
import VSpace from "components/VSpace";
import clsx from "clsx";

const useStyles = makeStyles((theme) => ({
  paper: {
    padding: theme.spacing(3),
  },
  paperTitle: {
    marginBottom: theme.spacing(1),
  },
  container: {
    overflowX: "auto",
  },
  active: {
    fontWeight: "bold",
  },
  button: {
    marginTop: theme.spacing(3),
  },
  dropdownButton: {
    backgroundColor: theme.palette.background.paper,
    "&:hover": {
      backgroundColor: theme.palette.secondary.main,
    },
    boxShadow: "none",
    color: theme.palette.common.white,
  },
}));

export interface BillingMethodsProps {
  organization: OrganizationByName_organizationByName_PrivateOrganization;
  billingInfo: BillingInfo_billingInfo;
  addCard: (value: boolean) => void;
}

const ViewBillingMethods: FC<BillingMethodsProps> = ({ organization, billingInfo, addCard }) => {
  const classes = useStyles();
  const { loading, error, data } = useQuery<BillingMethods, BillingMethodsVariables>(QUERY_BILLING_METHODS, {
    context: { ee: true },
    variables: { organizationID: organization.organizationID },
  });

  const [updateBillingMethod] = useMutation<UpdateBillingMethod, UpdateBillingMethodVariables>(UPDATE_BILLING_METHOD, {
    context: { ee: true },
  });

  if (!data || error) return null;

  const formatDriver = (driver: string) => {
    if (driver === STRIPECARD_DRIVER) return "Card";
    if (driver === STRIPEWIRE_DRIVER) return "Wire";
    if (driver === ANARCHISM_DRIVER) return "Anarchism";
  };

  const formatDetails = (driver: string, driverPayload: string) => {
    if (driver === STRIPECARD_DRIVER) {
      const payload = JSON.parse(driverPayload);
      const brand = payload.brand.toString();
      const last4 = payload.last4.toString();
      const expMonth = payload.expMonth.toString();
      const expYear = payload.expYear.toString();
      return `${brand.toUpperCase()} ${last4}, Exp: ${expMonth}/${expYear.substring(2, 4)}`;
    }
    if (driver === STRIPEWIRE_DRIVER) {
      const payload = JSON.parse(driverPayload);
      return `Invoices sent to ${payload.email_address}`;
    }
  };

  const getBillingMethodActions = (billingMethod: BillingInfo_billingInfo_billingMethod) => [
    {
      label: "Set to active",
      onClick: () =>
        updateBillingMethod({
          variables: {
            organizationID: organization.organizationID,
            billingMethodID: billingMethod.billingMethodID,
          },
        }),
    },
    // TODO: enable deleting billing methods
    // { label: "Delete billing method", onClick: () => deleteBillingMethod() }
  ];

  return (
    <>
      <Paper variant="outlined" className={classes.paper}>
        <Typography variant="h1" className={classes.paperTitle}>
          Billing methods
        </Typography>
        <Typography variant="body2" color="textSecondary">
          Payment information on file
        </Typography>
        <VSpace units={3} />
        {data.billingMethods.length === 0 && (
          // TODO: replace ContentContainer with just a nicely formatted callToAction for TitledPaper components
          <ContentContainer
            callToAction={{
              message: `You have no billing methods on file`,
              buttons: [{ label: "Add a credit card", onClick: () => addCard(true) }],
            }}
          />
        )}
        {data.billingMethods.length !== 0 && (
          <>
            <Grid container className={classes.container}>
              <Grid item xs={12}>
                <MuiTable>
                  <MuiTableHead>
                    <MuiTableRow>
                      <MuiTableCell>Type</MuiTableCell>
                      <MuiTableCell>Details</MuiTableCell>
                      <MuiTableCell>Active</MuiTableCell>
                      <MuiTableCell></MuiTableCell>
                    </MuiTableRow>
                  </MuiTableHead>
                  <MuiTableBody>
                    {data.billingMethods.map((billingMethod) => (
                      <MuiTableRow key={billingMethod.billingMethodID}>
                        <MuiTableCell
                          className={clsx(
                            billingMethod.billingMethodID === billingInfo.billingMethod?.billingMethodID &&
                              classes.active
                          )}
                        >
                          {formatDriver(billingMethod.paymentsDriver)}
                        </MuiTableCell>
                        <MuiTableCell
                          className={clsx(
                            billingMethod.billingMethodID === billingInfo.billingMethod?.billingMethodID &&
                              classes.active
                          )}
                        >
                          {formatDetails(billingMethod.paymentsDriver, billingMethod.driverPayload)}
                        </MuiTableCell>
                        <MuiTableCell
                          className={clsx(
                            billingMethod.billingMethodID === billingInfo.billingMethod?.billingMethodID &&
                              classes.active
                          )}
                        >
                          {billingMethod.billingMethodID === billingInfo.billingMethod?.billingMethodID ? "Yes" : "No"}
                        </MuiTableCell>
                        <MuiTableCell>
                          <DropdownButton
                            variant="contained"
                            margin="dense"
                            actions={getBillingMethodActions(billingMethod)}
                            className={classes.dropdownButton}
                          >
                            <MoreVert />
                          </DropdownButton>
                        </MuiTableCell>
                      </MuiTableRow>
                    ))}
                  </MuiTableBody>
                </MuiTable>
              </Grid>
            </Grid>
            <Button variant="contained" onClick={() => addCard(true)} className={classes.button}>
              Add card
            </Button>
          </>
        )}
      </Paper>
    </>
  );
};

export default ViewBillingMethods;
