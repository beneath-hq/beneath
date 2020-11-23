import { Grid, makeStyles, Paper, Typography } from "@material-ui/core";
import VSpace from "components/VSpace";
import { STRIPECARD_DRIVER } from "ee/lib/billing";
import React, { FC } from "react";

const useStyles = makeStyles((theme) => ({
  paperPadding: {
    padding: theme.spacing(3)
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
      <Paper className={classes.paperPadding} variant="outlined">
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
                <VSpace units={1} />
              </React.Fragment>
            ))}
          </>
        )}
      </Paper>
    </>
  );
};

export default ViewBillingMethod;