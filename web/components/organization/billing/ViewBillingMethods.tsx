import { useQuery } from "@apollo/react-hooks";
import _ from "lodash";
import React, { FC } from "react";
import useMe from "../../../hooks/useMe";

import { QUERY_BILLING_METHODS } from "../../../apollo/queries/billingmethod";
import { BillingMethods, BillingMethodsVariables } from "../../../apollo/types/BillingMethods";

import { Button, Dialog, DialogActions, DialogContent, DialogTitle, Grid, Typography } from "@material-ui/core";
import { makeStyles } from "@material-ui/core/styles";

import billing from "../../../lib/billing";
import CardForm from "./CardForm";

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

interface Props {
  organizationID: string;
}

const ViewBillingMethods: FC<Props> = ({ organizationID }) => {
  const classes = useStyles();
  const [addCardDialogue, setAddCardDialogue] = React.useState(false);

  const me = useMe();
  if (!me) {
    return <p>Need to log in to see your current billing plan</p>;
  }

  const { loading, error, data } = useQuery<BillingMethods, BillingMethodsVariables>(QUERY_BILLING_METHODS, {
    variables: {
      organizationID,
    },
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
          <Typography variant="h6" className={classes.title}>
            Billing methods on file
          </Typography>
        </Grid>
        <Grid item container direction="column">
          {cards && cards.map(({ billingMethodID, driverPayload }) => {
            const payload = JSON.parse(driverPayload);
            const rows = [
              { name: "Card type", detail: _.startCase(_.toLower(payload.brand)) },
              { name: "Card number", detail: "xxxx-xxxx-xxxx-" + payload.last4 },
              { name: "Expiration",
              detail: payload.expMonth.toString() + "/" + payload.expYear.toString().substring(2, 4) },
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
            onClick={() => { setAddCardDialogue(true); }}
          >
            Add Credit Card
          </Button>
          <Dialog
            open={addCardDialogue}
            fullWidth={true}
            maxWidth={"md"}
            onBackdropClick={() => { setAddCardDialogue(false); }}
          >
            <DialogTitle id="alert-dialog-title">{"Add a credit card"}</DialogTitle>
            <DialogContent>
              <CardForm closeDialogue={handleCloseDialogue}/>
            </DialogContent>
            <DialogActions />
          </Dialog>
        </Grid>
      </Grid>
    </React.Fragment>
  );
};

export default ViewBillingMethods;
