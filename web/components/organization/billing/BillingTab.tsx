import React, { FC } from 'react'
import _ from 'lodash'

import { Typography, Grid, Link, Paper } from "@material-ui/core"
import { makeStyles } from "@material-ui/core/styles"
import ViewBillingMethods from './ViewBillingMethods'
import ViewBillingInfo from './ViewBillingInfo'

const useStyles = makeStyles((theme) => ({
  banner: {
    padding: theme.spacing(2),
  },
}))

interface Props {
  organizationID: string
}

// TODO: create organization with Enterprise plan
const BillingTab: FC<Props> = ({ organizationID }) => {
  const classes = useStyles()

  return (
    <React.Fragment>
      <Paper elevation={1} square>
        <Typography className={classes.banner}>
          You can find detailed information about our billing plans at <Link href="https://about.beneath.dev/enterprise">about.beneath.dev/enterprise</Link>.
        </Typography>
      </Paper>

      <Grid container direction="column">
        <Grid item>
          <ViewBillingInfo organizationID={organizationID} />
        </Grid>
        <Grid item>
          <ViewBillingMethods organizationID={organizationID} />
        </Grid>
      </Grid>
    </React.Fragment>
  )
};

export default BillingTab;