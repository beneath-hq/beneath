import React, { FC } from 'react';
import { Button, Typography, Grid, Card, CardActions, CardContent, Box } from "@material-ui/core";
import { makeStyles } from "@material-ui/core/styles"

import billing from "../../../lib/billing"
import CardForm from "./driver/CardForm"

const useStyles = makeStyles((theme) => ({
  card: {
    minWidth: 325,
  },
  planName: {
    marginBottom: theme.spacing(2),
  },
  button: {
    marginTop: theme.spacing(.5),
    marginLeft: theme.spacing(.5),
    marginBottom: theme.spacing(.5)
  },
}))

const BillingPlanMenu: FC = () => {
  const [buyNow, setBuyNow] = React.useState(false)
  const [billingPlanToBuy, setBillingPlanToBuy] = React.useState("")
  const [contactUs, setContactUs] = React.useState(false)
  const classes = useStyles()

  if (buyNow) {
    return <CardForm billingPlanID={billingPlanToBuy}/>
  }

  if (contactUs) {
    // TODO: go to about.beneath.com/contact/demo
  }

  return (
    <React.Fragment>
      <Grid container direction="column" spacing={3}>
        <Grid item>
          <Typography align="center">Consider upgrading to one of our premium plans.</Typography>
        </Grid>
        <Grid item>
          <Typography align="center">Find detailed descriptions at about.beneath.network/pricing.</Typography>
        </Grid>
        <Grid container item spacing={3} alignItems="center" justify="center">
          <Grid item>
            <Card className={classes.card}>
              <CardContent>
                <Typography variant="h5" gutterBottom className={classes.planName}>
                  Professional plan
                </Typography>
                <Grid container direction="column" spacing={1}>
                  <Grid item>
                    <Typography variant="body2">
                      <Box fontWeight="fontWeightBold" component="span">
                        $50 per month base
                      </Box>
                    </Typography>
                  </Grid>
                  <Grid item>
                    <Typography variant="body2">
                      Base includes 25 GB reads and 5 GB writes
                    </Typography>
                  </Grid>
                  <Grid item>
                    <Typography variant="body2">
                      Private projects
                    </Typography>
                  </Grid>
                  <Grid item>
                    <Typography variant="body2">
                      Role-based access controls
                    </Typography>
                  </Grid>
                </Grid>
              </CardContent>
              <CardActions>
                <Button
                  variant="contained"
                  color="primary"
                  className={classes.button}
                  onClick={() => {
                    setBillingPlanToBuy(billing.PRO_MONTHLY_BILLING_PLAN_ID)
                    setBuyNow(true)
                  }}>
                  Buy Now
              </Button>
              </CardActions>
            </Card>
          </Grid>
          <Grid item>
            <Card className={classes.card}>
              <CardContent>
                <Typography variant="h5" gutterBottom className={classes.planName}>
                  Enterprise plan
                </Typography>
                <Grid container direction="column" spacing={1}>
                  <Grid item>
                    <Typography variant="body2">
                      <Box fontWeight="fontWeightBold" component="span">
                        Custom pricing
                      </Box>
                    </Typography>
                  </Grid>
                  <Grid item>
                    <Typography variant="body2">
                      Custom quotas by seat
                    </Typography>
                  </Grid>
                  <Grid item>
                    <Typography variant="body2">
                      Option to pay by wire
                    </Typography>
                  </Grid>
                  <Grid item>
                    <Typography variant="body2">
                      Premium support
                    </Typography>
                  </Grid>
                </Grid>
              </CardContent>
              <CardActions>
                <Button
                  variant="contained"
                  color="primary"
                  className={classes.button}
                  onClick={() => {
                    setContactUs(true)
                  }}>
                  Contact Us
                </Button>
              </CardActions>
            </Card>
          </Grid>
        </Grid>
      </Grid>
    </React.Fragment>
  )
}

export default BillingPlanMenu;