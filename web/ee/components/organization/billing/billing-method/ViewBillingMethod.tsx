import { Grid, makeStyles, Paper, Typography } from "@material-ui/core";
import VSpace from "components/VSpace";
import { STRIPECARD_DRIVER } from "ee/lib/billing";
import React, { FC } from "react";

const useStyles = makeStyles((theme) => ({
  paper: {
    padding: theme.spacing(3),
    height: "100%"
  },
  paperTitle: {
    marginBottom: theme.spacing(1),
  },
  textData: {
    fontWeight: "bold",
  }
}));

interface Props {
  paymentsDriver: string;
  driverPayload: string;
}

const ViewBillingMethod: FC<Props> = ({paymentsDriver, driverPayload}) => {
  const classes = useStyles();

  let brand: string = "";
  let exp: string = "";
  let last4: string = "";
  let rows;

  if (paymentsDriver === STRIPECARD_DRIVER) {
    const payload = JSON.parse(driverPayload);
    brand = payload.brand.toString().toUpperCase();
    last4 = payload.last4.toString();
    exp = `${payload.expMonth.toString()} / ${payload.expYear.toString().substring(2,4)}`;
    rows = [
      {key: "Brand", value: brand.toUpperCase()},
      {key: "Last 4 digits", value: last4},
      {key: "Expiration", value: exp}
    ];
  }


  return (
    <>
      <Paper variant="outlined" className={classes.paper}>
        <Typography variant="h2" className={classes.paperTitle}>
          Billing method
        </Typography>
        <Typography variant="body2" color="textSecondary">
          Your active billing method to be used for future payments
        </Typography>
        <VSpace units={3} />
        {paymentsDriver === STRIPECARD_DRIVER && rows && (
          <>
            {rows.map((row) => (
              <React.Fragment key={row.key}>
                <Grid container alignItems="center" spacing={1}>
                  <Grid item>
                    <Typography>
                      {row.key}:
                    </Typography>
                  </Grid>
                  <Grid item>
                    <Typography className={classes.textData}>
                      {row.value}
                    </Typography>
                  </Grid>
                </Grid>
              </React.Fragment>
            ))}
          </>
        )}
      </Paper>
    </>
  );
};

export default ViewBillingMethod;