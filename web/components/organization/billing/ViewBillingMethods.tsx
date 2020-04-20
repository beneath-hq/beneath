import React, { FC } from 'react'
import { useQuery } from "@apollo/react-hooks";
import useMe from "../../../hooks/useMe";
import _ from 'lodash'

import { QUERY_BILLING_METHODS } from '../../../apollo/queries/billingmethod';
import { BillingMethods, BillingMethodsVariables } from '../../../apollo/types/BillingMethods';

import { Button, Typography, Grid, Dialog, DialogActions, DialogTitle, DialogContent } from "@material-ui/core"
import { makeStyles } from "@material-ui/core/styles"

import CardForm from './CardForm';
import billing from "../../../lib/billing"

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
}))

interface Props {
  organizationID: string
}

const ViewBillingMethods: FC<Props> = ({ organizationID }) => {
  const classes = useStyles()
  const [addCardDialogue, setAddCardDialogue] = React.useState(false)

  const me = useMe();
  if (!me) {
    return <p>Need to log in to see your current billing plan</p>
  }

  const { loading, error, data } = useQuery<BillingMethods, BillingMethodsVariables>(QUERY_BILLING_METHODS, {
    variables: {
      organizationID: organizationID,
    },
  });


  if (error || !data) {
    return <p>Error: {JSON.stringify(error)}</p>;
  }

  const cards = data.billingMethods.filter(billingMethod => billingMethod.paymentsDriver == billing.STRIPECARD_DRIVER)
  const wire = data.billingMethods.filter(billingMethod => billingMethod.paymentsDriver == billing.STRIPEWIRE_DRIVER)[0]

  return (
    <React.Fragment>
      <Grid item container direction="column" xs={12} sm={6}>
        <Grid item>
          <Typography variant="h6" className={classes.title}>
            Billing methods on file
          </Typography>
        </Grid>
        <Grid item container direction="column">
          {cards.map(({ billingMethodID, driverPayload }) => {
            const payload = JSON.parse(driverPayload)
            const rows = [
              { name: 'Card type', detail: _.startCase(_.toLower(payload.brand)) },
              { name: 'Card number', detail: 'xxxx-xxxx-xxxx-' + payload.last4 },
              { name: 'Expiration', detail: payload.expMonth.toString() + '/' + payload.expYear.toString().substring(2, 4) },
            ]

            return (
              <React.Fragment key={billingMethodID}>
                <Grid item className={classes.billingMethod}>
                  {rows.map(rows => (
                    <React.Fragment key={rows.name}>
                      <Grid container>
                        <Grid item xs={12} sm={6}>
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
            )
          })}

          {wire && (
            <Grid item className={classes.billingMethod}>
              <Typography gutterBottom>You're authorized to pay by wire. Wire payments must be received within 15 days of the invoice.</Typography>
            </Grid>
          )}

          {cards.length == 0 && !wire && (
            <Grid item className={classes.billingMethod}>
              <Typography>You have no billing methods on file.</Typography>
            </Grid>
          )}
        </Grid>
        <Grid item>
          <Button
            className={classes.button}
            color="primary"
            onClick={() => { setAddCardDialogue(true) }}
          >
            Add Credit Card
          </Button>
          <Dialog
            open={addCardDialogue}
            fullWidth={true}
            maxWidth={"md"}
            onBackdropClick={() => { setAddCardDialogue(false) }}
          >
            <DialogTitle id="alert-dialog-title">{"Add a credit card"}</DialogTitle>
            <DialogContent>
              <CardForm />
            </DialogContent>
            <DialogActions />
          </Dialog>
        </Grid>
      </Grid>
    </React.Fragment>
  )
};

export default ViewBillingMethods;