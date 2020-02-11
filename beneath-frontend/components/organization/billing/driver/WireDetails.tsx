import React, { FC } from 'react';
import { makeStyles } from "@material-ui/core/styles"
import { Typography, Grid } from "@material-ui/core"

const useStyles = makeStyles((theme) => ({
  title: {
    marginTop: theme.spacing(2),
    marginBottom: theme.spacing(2),
  },
}))

const WireDetails: FC = () => {
  const classes = useStyles()

  return (
    <React.Fragment>
      <Grid item container direction="column" xs={12} sm={6}>
        <Typography variant="h6" className={classes.title}>
          Payment details
        </Typography>
        <Typography gutterBottom>You're paying by wire. Please make your payment within 15 days of receiving your invoice.</Typography>
      </Grid>
    </React.Fragment>
  )
}

export default WireDetails