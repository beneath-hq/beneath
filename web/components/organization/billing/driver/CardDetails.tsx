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

  // const OLDCardDetails: FC<Props> = ( { }) => {
  //   const [paymentDetails, setPaymentDetails] = React.useState<CardPaymentDetails | null>(null)
  //   const [editCard, setEditCard] = React.useState(false)
  //   const [error, setError] = React.useState("")
  //   const token = useToken()
  //   const classes = useStyles()
  
  //   // get current billing method 
  //   useEffect(() => {
  //     let isMounted = true
      
  //     const fetchData = async () => {
  //       const headers = { authorization: `Bearer ${token}` }
  //       let url = `${connection.API_URL}/billing/stripecard/get_payment_details`
  //       const res = await fetch(url, { headers })
  
  //       if (isMounted) {
  //         if (!res.ok) {
  //           setError(res.statusText)
  //           return
  //         }
  
  //         const paymentDetails: CardPaymentDetails = await res.json()
  //         setPaymentDetails(paymentDetails)
  //       }
  //     }
  
  //     fetchData()
  
  //     // avoid memory leak when component unmounts
  //     return () => {
  //       isMounted = false
  //     }
  //   }, [])
  
  //   if (error) {
  //     return <p>{error}</p>
  //   }
  
  //   // Q: is this a problem that it gets hits twice upon page reload? doesn't seem to have a visual effect.
  //   // I don't think this leads to flicker, because the other user profile tabs have a flicker too, even when Billing is not present
  //   if (paymentDetails == null) {
  //     // console.log("paymentDetails is null")
  //     return <p></p>
  //   }
  
  //   if (editCard) {
  //     return <CardForm />
  //   }
  
  //   const address = [paymentDetails.data.billingDetails.Address.Line1,
  //     paymentDetails.data.billingDetails.Address.Line2,
  //     paymentDetails.data.billingDetails.Address.City,
  //     paymentDetails.data.billingDetails.Address.State,
  //     paymentDetails.data.billingDetails.Address.PostalCode,
  //     paymentDetails.data.billingDetails.Address.Country].filter(Boolean) // omit Line2 if it's empty
  
  //   const payments = [
  //     { name: 'Card type', detail: _.startCase(_.toLower(paymentDetails.data.card.Brand))},
  //     { name: 'Card number', detail: 'xxxx-xxxx-xxxx-' + paymentDetails.data.card.Last4 },
  //     { name: 'Expiration', detail: paymentDetails.data.card.ExpMonth.toString() + '/' + paymentDetails.data.card.ExpYear.toString().substring(2,4) },
  //     { name: 'Card holder', detail: paymentDetails.data.billingDetails.Name },
  //     { name: 'Billing address', detail: address.join(', ')}
  //   ]
  
  //   return (
  //     <React.Fragment>
  //       <Grid container spacing={2}>
  //         <Grid item container direction="column" xs={12} sm={6}>
  //           <Grid container alignItems="center" justify="space-between">
  //             <Grid item>
  //               <Button
  //                 color="primary"
  //                 onClick={() => {
  //                   setEditCard(true)
  //                 }}>
  //                 Edit
  //             </Button>
  //             </Grid>
  //           </Grid>
  //           <Grid container>
  //             {payments.map(payments => (
  //               <React.Fragment key={payments.name}>
  //                 <Grid item xs={6}>
  //                   <Typography gutterBottom>{payments.name}</Typography>
  //                 </Grid>
  //                 <Grid item xs={6}>
  //                   <Typography gutterBottom>{payments.detail}</Typography>
  //                 </Grid>
  //               </React.Fragment>
  //             ))}
  //           </Grid>
  //         </Grid>
  //       </Grid>
  //     </React.Fragment>
  //   )
  // }