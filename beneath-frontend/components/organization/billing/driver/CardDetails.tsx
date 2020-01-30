import React, { FC, useEffect } from 'react'
import { Typography, Button, Grid } from "@material-ui/core"
import { makeStyles } from "@material-ui/core/styles"
import _ from 'lodash'

import { useToken } from '../../../../hooks/useToken'
import connection from "../../../../lib/connection"
import Loading from "../../../Loading"
import CardForm from './CardForm'
import CurrentBillingPlan from '../CurrentBillingPlan'

const useStyles = makeStyles((theme) => ({
  title: {
    marginTop: theme.spacing(2),
    marginBottom: theme.spacing(2),
  },
  loading: {
    marginTop: theme.spacing(5)
  }
}))

interface CardPaymentDetails {
  data: {
    organizationID: string,
    card: {
      Brand: string,
      Last4: string,
      ExpMonth: number,
      ExpYear: number,
    },
    billingDetails: {
      Name: string,
      Email: string,
      Address: {
        Line1: string,
        Line2: string,
        City: string,
        State: string,
        PostalCode: string,
        Country: string,
      }
    }
  },
  error: string | undefined
}

interface Props {
  billingPlanID: string
  billingPeriod: string
  description: string | null
}

const CardDetails: FC<Props> = ({ billingPlanID, billingPeriod, description }) => {
  const [paymentDetails, setPaymentDetails] = React.useState<CardPaymentDetails | null>(null)
  const [editCard, setEditCard] = React.useState(false)
  const [loading, setLoading] = React.useState(false)
  const [error, setError] = React.useState("")
  const token = useToken()
  const classes = useStyles()

  // get current payment details 
  useEffect(() => {
    let isMounted = true
    
    const fetchData = async () => {
      setLoading(true)
      const headers = { authorization: `Bearer ${token}` }
      let url = `${connection.API_URL}/billing/stripecard/get_payment_details`
      const res = await fetch(url, { headers })

      if (isMounted) {
        if (!res.ok) {
          setError(res.statusText)
          setLoading(false)
          return
        }

        const paymentDetails: CardPaymentDetails = await res.json()
        setPaymentDetails(paymentDetails)
        setLoading(false)
      }
    }

    fetchData()

    // avoid memory leak when component unmounts
    return () => {
      isMounted = false
    }
  }, [])

  if (loading) {
    return (
      <React.Fragment>
        <Grid item container direction="column" xs={12} sm={6}>
          <Grid item container alignItems="center" justify="space-between">
            <Grid item>
              <Typography variant="h6" className={classes.title}>
                Payment details
              </Typography>
            </Grid>
          </Grid>
          <Grid item container className={classes.loading}>
            <Loading justify="center" />
          </Grid>
        </Grid>
      </React.Fragment>
    )
  }

  if (error) {
    return <p>{error}</p>
  }

  // TODO: analyze why this is getting hit / whether it leads to a flicker
  if (paymentDetails == null) {
    console.log("paymentDetails is null")
    return <p></p>
  }

  if (editCard) {
    return <CardForm billingPlanID={billingPlanID} />
  }

  const address = [paymentDetails.data.billingDetails.Address.Line1,
    paymentDetails.data.billingDetails.Address.Line2,
    paymentDetails.data.billingDetails.Address.City,
    paymentDetails.data.billingDetails.Address.State,
    paymentDetails.data.billingDetails.Address.PostalCode,
    paymentDetails.data.billingDetails.Address.Country].filter(Boolean) // omit Line2 if it's empty

  const payments = [
    { name: 'Card type', detail: _.startCase(_.toLower(paymentDetails.data.card.Brand))},
    { name: 'Card number', detail: 'xxxx-xxxx-xxxx-' + paymentDetails.data.card.Last4 },
    { name: 'Expiration', detail: paymentDetails.data.card.ExpMonth.toString() + '/' + paymentDetails.data.card.ExpYear.toString().substring(2,4) },
    { name: 'Card holder', detail: paymentDetails.data.billingDetails.Name },
    { name: 'Billing address', detail: address.join(', ')}
  ]

  return (
    <React.Fragment>
      <Grid container spacing={2}>
        <CurrentBillingPlan billingPeriod={billingPeriod} description={description} />
        <Grid item container direction="column" xs={12} sm={6}>
          <Grid container alignItems="center" justify="space-between">
            <Grid item>
              <Typography variant="h6" className={classes.title}>
                Payment details
            </Typography>
            </Grid>
            <Grid item>
              <Button
                color="primary"
                onClick={() => {
                  setEditCard(true)
                }}>
                Edit
            </Button>
            </Grid>
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