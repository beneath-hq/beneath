import numbro from "numbro";
import { FC } from "react";

import { Box, Grid, LinearProgress, makeStyles, Theme, Typography } from "@material-ui/core";

const useStyles = makeStyles((theme: Theme) => ({
  title: {
    fontSize: "1.1rem",
    fontWeight: 400,
    paddingBottom: "4px",
  },
  usage: {
    fontSize: "1.5rem",
    fontWeight: 600,
    paddingBottom: "4px",
  },
  usageDenominator: {
    color: theme.palette.text.secondary,
    fontSize: theme.typography.body2.fontSize,
    fontWeight: 400,
  },
  percentageLabel: {
    marginLeft: theme.spacing(1),
    lineHeight: "100%",
  },
}));

export interface QuotaUsageIndicatorProps {
  label: string;
  dense?: boolean;
  usage: number;
  quota?: number | null;
  prepaidQuota?: number | null;
}

const percentFormat: numbro.Format = { output: "percent", mantissa: 0 };
const bytesFormat: numbro.Format = { base: "decimal", mantissa: 0, output: "byte" };

export const QuotaUsageIndicator: FC<QuotaUsageIndicatorProps> = ({ label, dense, usage, quota, prepaidQuota }) => {
  const progressQuota = !quota ? undefined : !prepaidQuota ? quota : usage >= prepaidQuota ? quota : prepaidQuota;

  const classes = useStyles();
  return (
    <Box>
      <Grid container alignItems="center">
        <Grid item xs={dense ? true : 12}>
          <Typography className={classes.title} noWrap>{label}</Typography>
        </Grid>
        <Grid item>
          <Typography className={classes.usage} noWrap>
            {numbro(usage).format(bytesFormat)}
            {quota &&
              (prepaidQuota && prepaidQuota !== quota ? (
                <span className={classes.usageDenominator}>
                  /{numbro(prepaidQuota).format(bytesFormat)} prepaid ({numbro(quota).format(bytesFormat)} cap)
                </span>
              ) : (
                <span className={classes.usageDenominator}>/{numbro(quota).format(bytesFormat)}</span>
              ))}
          </Typography>
        </Grid>
      </Grid>
      {progressQuota && (
        <Grid container wrap="nowrap" alignItems="center">
          <Grid item xs>
            <LinearProgress variant="determinate" value={Math.min((usage / progressQuota) * 100, 100)} />
          </Grid>
          <Grid item>
            <Typography className={classes.percentageLabel} variant="overline">
              {numbro(usage / progressQuota).format(percentFormat)}
            </Typography>
          </Grid>
        </Grid>
      )}
    </Box>
  );
};

export default QuotaUsageIndicator;
