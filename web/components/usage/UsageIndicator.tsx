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

interface IndicatorProps {
  label: string;
  dense?: boolean;
  children?: any;
}

const intFormat: numbro.Format = { thousandSeparated: true };
const bytesFormat: numbro.Format = { base: "decimal", mantissa: 0, output: "byte" };
const percentFormat: numbro.Format = { output: "percent", mantissa: 0 };

const Indicator: FC<IndicatorProps> = ({ label, dense, children }) => {
  const classes = useStyles();
  return (
    <Grid container alignItems="center">
      {!dense && (
        <Grid item xs={12}>
          <Typography className={classes.title} noWrap>
            {label}
          </Typography>
        </Grid>
      )}
      <Grid item xs={dense}>
        <Typography className={classes.usage} noWrap>
          {children}
        </Typography>
      </Grid>
      {dense && (
        <Grid item>
          <Typography className={classes.title} noWrap>
            {label}
          </Typography>
        </Grid>
      )}
    </Grid>
  );
};

export interface UsageIndicatorProps extends IndicatorProps {
  usage: number;
  format?: "integer" | "bytes";
}

export const UsageIndicator: FC<UsageIndicatorProps> = ({ format, usage, ...indicatorProps }) => {
  return (
    <Indicator {...indicatorProps}>
      {format === "bytes" ? numbro(usage).format(bytesFormat) : numbro(usage).format(intFormat)}
    </Indicator>
  );
};

export interface QuotaUsageIndicatorProps extends IndicatorProps {
  usage: number;
  quota?: number | null;
  prepaidQuota?: number | null;
}

export const QuotaUsageIndicator: FC<QuotaUsageIndicatorProps> = ({
  usage,
  quota,
  prepaidQuota,
  ...indicatorProps
}) => {
  const progressQuota = !quota ? undefined : !prepaidQuota ? quota : usage >= prepaidQuota ? quota : prepaidQuota;

  const classes = useStyles();
  return (
    <Box>
      <Indicator {...indicatorProps}>
        {numbro(usage).format(bytesFormat)}
        {quota &&
          (prepaidQuota && prepaidQuota !== quota ? (
            <span className={classes.usageDenominator}>
              /{numbro(prepaidQuota).format(bytesFormat)} prepaid ({numbro(quota).format(bytesFormat)} cap)
            </span>
          ) : (
            <span className={classes.usageDenominator}>/{numbro(quota).format(bytesFormat)}</span>
          ))}
      </Indicator>
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
