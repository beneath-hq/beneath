import numbro from "numbro";
import { FC } from "react";

import { Grid, LinearProgress, makeStyles, Paper, Theme, Typography } from "@material-ui/core";

const useStyles = makeStyles((theme: Theme) => ({
  paper: {
    padding: theme.spacing(2),
  },
  percentageLabel: {
    marginLeft: theme.spacing(1),
    lineHeight: "100%",
  },
  usageLabel: {
    lineHeight: "100%",
    textAlign: "right",
  },
}));

export interface UsageIndicatorProps {
  standalone: boolean;
  kind: "read" | "write" | "scan";
  usage: number;
  quota: number;
}

export const UsageIndicator: FC<UsageIndicatorProps> = ({ standalone, kind, usage, quota }) => {
  const classes = useStyles();
  if (standalone) {
    return (
      <Grid container item xs={12} md={6}>
        <Grid item xs={12}>
          <Paper className={classes.paper}>
            <Typography variant="overline">{kind === "read" ? "Read quota usage" : kind === "write" ? "Write quota usage" : "SQL scan quota usage"}</Typography>
            <ActualIndicator standalone={standalone} kind={kind} usage={usage} quota={quota} />
          </Paper>
        </Grid>
      </Grid>
    );
  }

  return <ActualIndicator standalone={standalone} kind={kind} usage={usage} quota={quota} />;
};

export default UsageIndicator;

const percentFormat: numbro.Format = { output: "percent", mantissa: 0 };
const bytesFormat: numbro.Format = { base: "decimal", mantissa: 0, output: "byte" };

export const ActualIndicator: FC<UsageIndicatorProps> = ({ standalone, kind, usage, quota }) => {
  const classes = useStyles();
  return (
    <Grid container item xs={12}>
      <Grid container item xs={12} wrap="nowrap" alignItems="center">
        <Grid item xs>
          <LinearProgress variant="determinate" value={Math.min((usage / quota) * 100, 100)} />
        </Grid>
        <Grid item>
          <Typography className={classes.percentageLabel} variant="overline">
            {numbro(usage / quota).format(percentFormat)}
          </Typography>
        </Grid>
      </Grid>
      <Grid item xs={12}>
        <Typography variant="overline" className={classes.usageLabel}>
          Used {numbro(usage).format(bytesFormat)} of {numbro(quota).format(bytesFormat)} this month
        </Typography>
      </Grid>
    </Grid>
  );
};
