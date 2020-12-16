import { Grid, makeStyles, Tab, Tabs, Theme, Typography } from "@material-ui/core";
import { FC, useState } from "react";
import Moment from "react-moment";

import { StreamInstance } from "./types";
import { EntityKind } from "apollo/types/globalTypes";
import { StreamByOrganizationProjectAndName_streamByOrganizationProjectAndName } from "apollo/types/StreamByOrganizationProjectAndName";
import ErrorNote from "components/ErrorNote";
import { useHourlyUsage, useTotalUsage } from "components/usage/hooks";
import { UsageIndicator } from "components/usage/UsageIndicator";
import { PaperGrid } from "components/Paper";
import UsageChart, { UsageUnit } from "components/usage/UsageChart";

const useStyles = makeStyles((theme: Theme) => ({
  tab: {
    fontSize: theme.typography.body2.fontSize,
    padding: theme.spacing(1.5),
  },
}));

export interface StreamUsageViewProps {
  stream: StreamByOrganizationProjectAndName_streamByOrganizationProjectAndName;
  instance: StreamInstance;
}

export const StreamUsageView: FC<StreamUsageViewProps> = ({ stream, instance }) => {
  const classes = useStyles();

  const units: UsageUnit[] = ["bytes", "ops", "records"];
  const [unit, setUnit] = useState(units[2]);

  const dimensions: ("Current" | "All")[] = ["Current", "All"];
  const [dimension, setDimension] = useState(dimensions[0]);

  const entityKind = dimension === "Current" ? EntityKind.StreamInstance : EntityKind.Stream;
  const entityID = dimension === "Current" ? instance.streamInstanceID : stream.streamID;

  const { data: totalUsage, loading: totalUsageLoading, error: totalUsageError } = useTotalUsage(entityKind, entityID);

  const { data: hourlyUsages, loading: hourlyUsageLoading, error: hourlyUsageError } = useHourlyUsage(
    entityKind,
    entityID
  );

  if (totalUsageError || hourlyUsageError) {
    return <ErrorNote error={totalUsageError || hourlyUsageError} />;
  }

  return (
    <Grid container spacing={2}>
      <Grid item xs={12}>
        <Grid container alignItems="center">
          <Grid item xs>
            <Typography variant="h2">Total usage</Typography>
            <Typography variant="subtitle2">
              {dimension === "Current" && (
                <>
                  Instance created on <Moment date={instance.createdOn} format="D MMM YYYY" />
                </>
              )}
              {dimension === "All" && (
                <>
                  Stream created on <Moment date={stream.createdOn} format="D MMM YYYY" />
                </>
              )}
            </Typography>
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
                  label={dimension + (dimension === "Current" ? " instance" : " instances")}
                  value={dimension}
                  className={classes.tab}
                />
              ))}
            </Tabs>
          </Grid>
        </Grid>
      </Grid>
      {totalUsage && (
        <>
          <PaperGrid item xs={12} md={4}>
            <UsageIndicator label="Records written" usage={totalUsage.writeRecords} />
          </PaperGrid>
          <PaperGrid item xs={12} md={4}>
            <UsageIndicator label="Bytes written" usage={totalUsage.writeBytes} format="bytes" />
          </PaperGrid>
          <PaperGrid item xs={12} md={4}>
            <UsageIndicator label="Read calls" usage={totalUsage.readOps} />
          </PaperGrid>
        </>
      )}
      {hourlyUsages && (
        <>
          <Grid item xs={12}>
            <Grid container alignItems="center">
              <Grid item xs>
                <Typography variant="h2">Recent writes</Typography>
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
          <PaperGrid item xs={12}>
            <UsageChart usages={hourlyUsages} dimension="write" unit={unit} />
          </PaperGrid>
        </>
      )}
    </Grid>
  );
};

export default StreamUsageView;
