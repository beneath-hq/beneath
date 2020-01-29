import React, { FC, useEffect } from 'react'
import { Typography, Button, Grid } from "@material-ui/core"
import { makeStyles } from "@material-ui/core/styles"
import _ from 'lodash'

import { useToken } from '../../../../hooks/useToken'
import connection from "../../../../lib/connection"
import Loading from "../../../Loading"
import CurrentBillingPlan from '../CurrentBillingPlan'
import CardFormStripe from './CardFormStripe'

const useStyles = makeStyles((theme) => ({
  title: {
    marginTop: theme.spacing(2),
    marginBottom: theme.spacing(2),
  },
}))

interface CardPaymentDetails {
  data: {
    organization_id: string,
    card: {
      Brand: string,
      Last4: string,
      ExpMonth: number,
      ExpYear: number,
    },
    billing_details: {
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
  description: string | null
  billing_period: string
  billing_plan_id: string
}

const CardDetails: FC<Props> = ({ description, billing_period, billing_plan_id }) => {
  const [paymentDetails, setPaymentDetails] = React.useState<CardPaymentDetails | null>(null)
  const [error, setError] = React.useState("")
  const [loading, setLoading] = React.useState(false)
  const [editCard, setEditCard] = React.useState(false)
  const token = useToken()
  const classes = useStyles()  

  // get current payment details 
  useEffect(() => {
    let isMounted = true
    
    const fetchData = async () => {
      setLoading(true)
      const headers = { authorization: `Bearer ${token}` }
      let payment_details_url = `${connection.API_URL}/billing/stripecard/get_payment_details`
      const res = await fetch(payment_details_url, { headers })

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
    return <Loading justify="center" />
  }

  if (error) {
    return <p>{error}</p>
  }

  // TODO: this is getting hit "before flicker"
  if (paymentDetails == null) {
    console.log("got here")
    return <p></p>
  }

  if (editCard) {
    return <CardFormStripe billing_plan_id={billing_plan_id} />
  }

  const address = [paymentDetails.data.billing_details.Address.Line1,
    paymentDetails.data.billing_details.Address.Line2,
    paymentDetails.data.billing_details.Address.City,
    paymentDetails.data.billing_details.Address.State,
    paymentDetails.data.billing_details.Address.PostalCode,
    paymentDetails.data.billing_details.Address.Country].filter(Boolean) // omit Line2 if it's empty

  const payments = [
    { name: 'Card type', detail: _.startCase(_.toLower(paymentDetails.data.card.Brand))},
    { name: 'Card number', detail: 'xxxx-xxxx-xxxx-' + paymentDetails.data.card.Last4 },
    { name: 'Expiration', detail: paymentDetails.data.card.ExpMonth.toString() + '/' + paymentDetails.data.card.ExpYear.toString().substring(2,4) },
    { name: 'Card holder', detail: paymentDetails.data.billing_details.Name },
    { name: 'Billing address', detail: address.join(', ')}
  ]

  // current card details
  const CardBillingDetails = (
    <React.Fragment>
      <Grid container spacing={2}>
        <CurrentBillingPlan description={description} billing_period={billing_period} />
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

  return CardBillingDetails
}

export default CardDetails