import { Grid, makeStyles, Tab, Tabs, Theme, Typography } from "@material-ui/core";
import { FC, useEffect, useState } from "react";
import Moment from "react-moment";

import { EntityKind } from "apollo/types/globalTypes";
import ErrorNote from "components/ErrorNote";
import { usageDescriptionFor, UsageDimension, UsageUnit, useHourlyUsage, useQuotaUsage } from "components/usage/util";
import { UsageIndicator, QuotaUsageIndicator } from "components/usage/UsageIndicator";
import { PaperGrid } from "components/Paper";
import Loading from "components/Loading";
import UsageChart from "./UsageChart";
import {  } from "./util";

const useStyles = makeStyles((theme: Theme) => ({
  tab: {
    fontSize: theme.typography.body2.fontSize,
    padding: theme.spacing(1.5),
  },
}));

export interface OwnerUsageViewProps {
  entityKind: EntityKind;
  entityID: string;
  quotaStartTime: string;
  quotaEndTime: string;
  readQuota?: number | null;
  writeQuota?: number | null;
  scanQuota?: number | null;
  prepaidReadQuota?: number | null;
  prepaidWriteQuota?: number | null;
  prepaidScanQuota?: number | null;
}

export const OwnerUsageView: FC<OwnerUsageViewProps> = (props) => {
  const classes = useStyles();

  const units: UsageUnit[] = ["bytes", "ops", "records"];
  const [unit, setUnit] = useState(units[0]);

  const dimensions: UsageDimension[] = ["read", "write", "scan"];
  const [dimension, setDimension] = useState(dimensions[0]);

  useEffect(() => {
    if (unit === "records" && dimension === "scan") {
      setDimension("read");
    }
  }, [unit]);

  const { data: quotaUsage, loading: quotaUsageLoading, error: quotaUsageError } = useQuotaUsage(
    props.entityKind,
    props.entityID,
    props.quotaStartTime
  );

  const { data: hourlyUsages, loading: hourlyUsageLoading, error: hourlyUsageError } = useHourlyUsage(
    props.entityKind,
    props.entityID
  );

  if (quotaUsageLoading || hourlyUsageLoading) {
    return <Loading justify="center" />;
  }

  if (quotaUsageError || hourlyUsageError || !quotaUsage || !hourlyUsages) {
    return <ErrorNote error={quotaUsageError || hourlyUsageError} />;
  }

  return (
    <Grid container spacing={2}>
      <Grid item xs={12}>
        <Grid container alignItems="center">
          <Grid item xs>
            <Typography variant="h2">Current usage</Typography>
            <Typography variant="subtitle2">
              {props.entityKind === EntityKind.Organization ? "Billing period" : "Period"} runs from{" "}
              <Moment date={props.quotaStartTime} format="D MMM" /> to{" "}
              <Moment date={props.quotaEndTime} format="D MMM" />
            </Typography>
          </Grid>
          <Grid item>
            <Tabs value={unit} onChange={(_, value) => setUnit(value)} variant="scrollable" scrollButtons="auto">
              {units.map((unit) => (
                <Tab
                  key={unit}
                  className={classes.tab}
                  value={unit}
                  label={unit === "bytes" ? "Bytes" : unit === "records" ? "Records" : "Requests"}
                />
              ))}
            </Tabs>
          </Grid>
        </Grid>
      </Grid>
      {unit === "bytes" && (
        <>
          <PaperGrid item xs={12} md={4}>
            <QuotaUsageIndicator
              label="Reads"
              usage={quotaUsage.readBytes}
              quota={props.readQuota}
              prepaidQuota={props.prepaidReadQuota}
            />
          </PaperGrid>
          <PaperGrid item xs={12} md={4}>
            <QuotaUsageIndicator
              label="Writes"
              usage={quotaUsage.writeBytes}
              quota={props.writeQuota}
              prepaidQuota={props.prepaidWriteQuota}
            />
          </PaperGrid>
          <PaperGrid item xs={12} md={4}>
            <QuotaUsageIndicator
              label="Scans"
              usage={quotaUsage?.scanBytes}
              quota={props.scanQuota}
              prepaidQuota={props.prepaidScanQuota}
            />
          </PaperGrid>
        </>
      )}
      {unit === "records" && (
        <>
          <PaperGrid item xs={12} md={6}>
            <UsageIndicator label="Records read" usage={quotaUsage.readRecords} />
          </PaperGrid>
          <PaperGrid item xs={12} md={6}>
            <UsageIndicator label="Records written" usage={quotaUsage.writeRecords} />
          </PaperGrid>
        </>
      )}
      {unit === "ops" && (
        <>
          <PaperGrid item xs={12} md={4}>
            <UsageIndicator label="Read requests" usage={quotaUsage.readOps} />
          </PaperGrid>
          <PaperGrid item xs={12} md={4}>
            <UsageIndicator label="Write requests" usage={quotaUsage.writeOps} />
          </PaperGrid>
          <PaperGrid item xs={12} md={4}>
            <UsageIndicator label="Scan requests" usage={quotaUsage.scanOps} />
          </PaperGrid>
        </>
      )}
      <Grid item xs={12}>
        <Grid container alignItems="center">
          <Grid item xs>
            <Typography variant="h2">Recent activity</Typography>
          </Grid>
          <Grid item>
            <Tabs
              value={dimension}
              onChange={(_, value) => setDimension(value)}
              variant="scrollable"
              scrollButtons="auto"
            >
              {dimensions.map((dimension) => (
                <Tab
                  key={dimension}
                  className={classes.tab}
                  value={dimension}
                  label={dimension === "read" ? "Reads" : dimension === "write" ? "Writes" : "Scans"}
                  disabled={dimension === "scan" && unit === "records"}
                />
              ))}
            </Tabs>
          </Grid>
        </Grid>
      </Grid>
      <PaperGrid item xs={12}>
        <UsageChart usages={hourlyUsages} unit={unit} dimension={dimension} />
      </PaperGrid>
    </Grid>
  );
};

export default OwnerUsageView;
