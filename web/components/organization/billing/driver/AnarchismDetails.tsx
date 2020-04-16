import React, { FC } from 'react';
import { makeStyles } from "@material-ui/core/styles"
import { Typography, Button, Grid } from "@material-ui/core"

const useStyles = makeStyles((theme) => ({
  title: {
    marginTop: theme.spacing(2),
    marginBottom: theme.spacing(2),
  },
  button: {
    marginTop: theme.spacing(3),
  },
}))

const AnarchismDetails: FC = () => {
  const classes = useStyles()

  // return (
  //   <React.Fragment>
  //     <Grid item container direction="column" xs={12} sm={6}>
  //       <Typography variant="h6" className={classes.title}>
  //         Billing method
  //       </Typography>
  //       <Typography gutterBottom>
  //         You don't have an active billing method.
  //       </Typography>
  //       <Grid item>
  //         <Button
  //           variant="outlined"
  //           color="primary"
  //           className={classes.button}
  //           onClick={() => {
  //             setAddCard(true)
  //           }}>
  //           Add Credit Card
  //         </Button>
  //       </Grid>
  //     </Grid>
  //   </React.Fragment>
  // )
  return (
    <React.Fragment>
      <Typography gutterBottom>
        You don't have an active billing method.
      </Typography>
    </React.Fragment>
  )
}

export default AnarchismDetails