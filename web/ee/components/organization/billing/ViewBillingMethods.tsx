import { useQuery } from "@apollo/client";
import _ from "lodash";
import dynamic from "next/dynamic";
import React, { FC } from "react";

import { QUERY_BILLING_METHODS } from "ee/apollo/queries/billingMethod";
import { BillingMethods, BillingMethodsVariables } from "ee/apollo/types/BillingMethods";
import { OrganizationByName_organizationByName_PrivateOrganization } from "apollo/types/OrganizationByName";
import billing from "ee/lib/billing";

import { Button, Dialog, DialogActions, DialogContent, DialogTitle, Grid, Typography } from "@material-ui/core";
import { makeStyles } from "@material-ui/core/styles";

const useStyles = makeStyles((theme) => ({
  title: {
    marginTop: theme.spacing(4),
    marginBottom: theme.spacing(2),
  },
  button: {
    marginTop: theme.spacing(3),
    marginBotton: theme.spacing(2),
    marginRight: theme.spacing(3),
  },
  billingMethod: {
    marginBottom: theme.spacing(1),
  },
}));

export interface BillingMethodsProps {
  organization: OrganizationByName_organizationByName_PrivateOrganization;
}

const ViewBillingMethods: FC<BillingMethodsProps> = ({ organization }) => {
  const classes = useStyles();
  const [addCardDialogue, setAddCardDialogue] = React.useState(false);
  const DynamicCardForm = dynamic(() => import("./CardForm"));

  const { loading, error, data } = useQuery<BillingMethods, BillingMethodsVariables>(QUERY_BILLING_METHODS, {
    context: { ee: true },
    variables: { organizationID: organization.organizationID },
  });

  if (error || !data) {
    return <p>Error: {JSON.stringify(error)}</p>;
  }

  let cards;
  let wire;
  if (data.billingMethods && data.billingMethods.length > 0) {
    cards = data.billingMethods.filter((billingMethod) => billingMethod.paymentsDriver === billing.STRIPECARD_DRIVER);
    wire = data.billingMethods.filter((billingMethod) => billingMethod.paymentsDriver === billing.STRIPEWIRE_DRIVER)[0];
  }

  const handleCloseDialogue = () => {
    setAddCardDialogue(false);
    return;
  };

  return (
    <React.Fragment>
      <Grid container direction="column">
        <Grid item>
          <Typography variant="h5" className={classes.title}>
            Billing methods on file
          </Typography>
        </Grid>
        <Grid item container direction="column">
          {cards &&
            cards.map(({ billingMethodID, driverPayload }) => {
              const payload = JSON.parse(driverPayload);
              const rows = [
                { name: "Card type", detail: _.startCase(_.toLower(payload.brand)) },
                { name: "Card number", detail: "xxxx-xxxx-xxxx-" + payload.last4 },
                {
                  name: "Expiration",
                  detail: payload.expMonth.toString() + "/" + payload.expYear.toString().substring(2, 4),
                },
              ];

              return (
                <React.Fragment key={billingMethodID}>
                  <Grid item className={classes.billingMethod}>
                    {rows.map((rows) => (
                      <React.Fragment key={rows.name}>
                        <Grid container>
                          <Grid item xs={6} sm={4} md={2}>
                            <Typography gutterBottom>{rows.name}</Typography>
                          </Grid>
                          <Grid item>
                            <Typography gutterBottom>{rows.detail}</Typography>
                          </Grid>
                        </Grid>
                      </React.Fragment>
                    ))}
                  </Grid>
                </React.Fragment>
              );
            })}

          {wire && (
            <Grid item className={classes.billingMethod}>
              <Typography gutterBottom>
                You're authorized to pay by wire. Wire payments must be received within 15 days of the invoice.
              </Typography>
            </Grid>
          )}

          {!cards && !wire && (
            <Grid item className={classes.billingMethod}>
              <Typography>None.</Typography>
            </Grid>
          )}
        </Grid>
        <Grid item>
          <Button
            className={classes.button}
            color="primary"
            onClick={() => {
              setAddCardDialogue(true);
            }}
          >
            Add Credit Card
          </Button>
          <Dialog
            open={addCardDialogue}
            fullWidth={true}
            maxWidth={"md"}
            onBackdropClick={() => {
              setAddCardDialogue(false);
            }}
          >
            <DialogTitle id="alert-dialog-title">{"Add a credit card"}</DialogTitle>
            <DialogContent>
              <DynamicCardForm organization={organization} closeDialogue={handleCloseDialogue} />
            </DialogContent>
            <DialogActions />
          </Dialog>
        </Grid>
      </Grid>
    </React.Fragment>
  );
};

export default ViewBillingMethods;
