import { Grid, makeStyles, Paper, Typography } from "@material-ui/core";
import VSpace from "components/VSpace";
import { STRIPECARD_DRIVER } from "ee/lib/billing";
import { FC } from "react";

const useStyles = makeStyles((theme) => ({
  paperPadding: {
    padding: theme.spacing(3)
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
  if (paymentsDriver === STRIPECARD_DRIVER) {
    const payload = JSON.parse(driverPayload);
    brand = payload.brand.toString().toUpperCase();
    last4 = payload.last4.toString();
    exp = `${payload.expMonth.toString()} / ${payload.expYear.toString().substring(2,4)}`;
  }

  return (
    <>
      <Paper className={classes.paperPadding} variant="outlined">
        {paymentsDriver === STRIPECARD_DRIVER && (
          <>
            <Grid container justify="space-between" alignItems="center">
              <Grid item>
                <Typography>
                  Brand:
                </Typography>
              </Grid>
              <Grid item>
                <Typography>
                  {brand.toUpperCase()}
                </Typography>
              </Grid>
            </Grid>
            <VSpace units={1} />
            <Grid container justify="space-between" alignItems="center">
              <Grid item>
                <Typography>
                  Last four digits:
                </Typography>
              </Grid>
              <Grid item>
                <Typography>
                  {last4}
                </Typography>
              </Grid>
            </Grid>
            <VSpace units={1} />
            <Grid container justify="space-between" alignItems="center">
              <Grid item>
                <Typography>
                  Expiration:
                </Typography>
              </Grid>
              <Grid item>
                <Typography>
                  {exp}
                </Typography>
              </Grid>
            </Grid>
          </>
        )}
      </Paper>
    </>
  );
};

export default ViewBillingMethod;