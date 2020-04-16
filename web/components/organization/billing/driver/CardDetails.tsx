import React, { FC } from 'react'
import { Typography, Button, Grid } from "@material-ui/core"
import { makeStyles } from "@material-ui/core/styles"
import _ from 'lodash'

import CardForm from './CardForm'

const useStyles = makeStyles((theme) => ({
  title: {
    marginTop: theme.spacing(2),
    marginBottom: theme.spacing(2),
  },
  loading: {
    marginTop: theme.spacing(5)
  }
}))


interface Props {
  billingMethodID: string,
  driverPayload: string,
}

const CardDetails: FC<Props> = ({ billingMethodID, driverPayload }) => {
  const [editCard, setEditCard] = React.useState(false)
  const classes = useStyles()

  const details = JSON.parse(driverPayload)

  if (editCard) {
    return <CardForm />
  }

  const payments = [
    { name: 'Card type', detail: _.startCase(_.toLower(details.brand))},
    { name: 'Card number', detail: 'xxxx-xxxx-xxxx-' + details.last4 },
    { name: 'Expiration', detail: details.expMonth.toString() + '/' + details.expYear.toString().substring(2,4) },
  ]

  return (
    <React.Fragment>
      <Grid container spacing={2}>
        <Grid item container direction="column" xs={12} sm={6}>
          <Grid container alignItems="center" justify="space-between">
            {/* <Grid item>
              <Button
                color="primary"
                onClick={() => {
                  setEditCard(true)
                }}>
                Edit
            </Button>
            </Grid> */}
          </Grid>
          <Grid container>
            {payments.map(payments => (
              <React.Fragment key={payments.name}>
                <Grid item xs={6}>
                  <Typography gutterBottom>{payments.name}</Typography>
                </Grid>
                <Grid item xs={6}>
                  <Typography gutterBottom>{payments.detail}</Typography>
                </Grid>
              </React.Fragment>
            ))}
          </Grid>
        </Grid>
      </Grid>
    </React.Fragment>
  )
}

export default CardDetails
