import React, { FC } from 'react';
import CurrentBillingPlan from '../CurrentBillingPlan'
import { makeStyles } from "@material-ui/core/styles"
import Grid from '@material-ui/core/Grid'
import { Typography } from "@material-ui/core"

const useStyles = makeStyles((theme) => ({
  title: {
    marginTop: theme.spacing(2),
    marginBottom: theme.spacing(2),
  },
}))

interface Props {
  description: string | null
  billing_period: string
}

const AnarchismDetails: FC<Props> = ({ description, billing_period }) => {
  const classes = useStyles()

  return (
    <React.Fragment>
      <Grid container spacing={2}>
        <CurrentBillingPlan description={description} billing_period={billing_period} />
        <Grid item container direction="column" xs={12} sm={6}>
          <Typography variant="h6" className={classes.title}>
            Payment details
          </Typography>
          <Typography gutterBottom>You are on a special plan... you don't have to pay!</Typography>
        </Grid>
      </Grid>
    </React.Fragment>
  )
}

export default AnarchismDetails