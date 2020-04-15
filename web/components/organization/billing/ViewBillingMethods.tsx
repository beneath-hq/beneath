import React, { FC } from 'react';
import useMe from "../../../hooks/useMe";
import { makeStyles } from "@material-ui/core/styles"
import { Typography, Grid, List, ListItem, ListItemText } from "@material-ui/core"
import CardDetails from "./driver/CardDetails"
import { BillingMethods, BillingMethodsVariables } from '../../../apollo/types/BillingMethods';
import { QUERY_BILLING_METHODS } from '../../../apollo/queries/billingmethod';
import { useQuery } from '@apollo/react-hooks';
import WireDetails from './driver/WireDetails';

const useStyles = makeStyles((theme) => ({
  title: {
    marginTop: theme.spacing(2),
    marginBottom: theme.spacing(2),
  },
  button: {
    marginTop: theme.spacing(3),
  },
  buttons: {
    marginTop: theme.spacing(3),
  },
}))

interface Props {
  organizationID: string
}

const ViewBillingMethods: FC<Props> = ( { organizationID }) => {
  const classes = useStyles()

  const me = useMe();
  if (!me) {
    return <p>Need to log in</p>
  }

  const { loading, error, data } = useQuery<BillingMethods, BillingMethodsVariables>(QUERY_BILLING_METHODS, {
    variables: {
      organizationID: organizationID,
    },
  });

  if (error || !data) {
    return <p>Error: {JSON.stringify(error)}</p>;
  }

  const cards = data.billingMethods.filter(billingMethod => billingMethod.paymentsDriver == "stripecard")
  const wire = data.billingMethods.filter(billingMethod => billingMethod.paymentsDriver == "stripewire")

  return (
    <React.Fragment>
      <Typography variant="h6" className={classes.title}>
        Billing methods
      </Typography>
      {cards.map(({ billingMethodID, driverPayload }) => ( 
        <CardDetails billingMethodID={billingMethodID} driverPayload={driverPayload} />
      ))}

      {wire.map(({ billingMethodID }) => (
        <WireDetails />
      ))}
      
      {cards.length == 0 && wire.length == 0 && (
        <Typography>
          You have no billing methods on file
        </Typography>
      )}
    </React.Fragment>
  )
}

export default ViewBillingMethods
