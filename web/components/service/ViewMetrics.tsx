import { Grid } from "@material-ui/core";
import { FC } from "react";

import { EntityKind } from "../../apollo/types/globalTypes";
import { ServiceByNameAndOrganization_serviceByNameAndOrganization } from "../../apollo/types/ServiceByNameAndOrganization";
import ErrorNote from "../ErrorNote";
import { useMonthlyMetrics, useWeeklyMetrics } from "../metrics/hooks";
import TopIndicators from "../metrics/user/TopIndicators";
import UsageIndicator from "../metrics/user/UsageIndicator";
import WeekChart from "../metrics/WeekChart";

export interface ViewMetricsProps {
  service: ServiceByNameAndOrganization_serviceByNameAndOrganization;
}

const ViewMetrics: FC<ViewMetricsProps> = ({ service }) => {
  const sharedProps = {
    entityKind: EntityKind.Service,
    entityID: service.serviceID,
  };

  const props = {
    readQuota: service.readQuota,
    writeQuota: service.writeQuota,
  };

  // readBytes
  // ;

  return (
    <>
      <Grid container spacing={2}>
        <MetricsOverview
          entityKind={EntityKind.Service}
          entityID={service.serviceID}
          readQuota={service.readQuota}
          writeQuota={service.writeQuota}
        />
        <MetricsWeek
          entityKind={EntityKind.Service}
          entityID={service.serviceID}
          field="readBytes"
          title="Data read in the last 7 days"
        />
        <MetricsWeek
          entityKind={EntityKind.Service}
          entityID={service.serviceID}
          field="writeBytes"
          title="Data written in the last 7 days"
        />
      </Grid>
    </>
  );
};

export default ViewMetrics;

export interface MetricsOverviewProps {
  entityKind: EntityKind;
  entityID: string;
  readQuota: number | null;
  writeQuota: number | null;
}

const MetricsOverview: FC<MetricsOverviewProps> = ({ entityKind, entityID, readQuota, writeQuota }) => {
  const { metrics, total, latest, error } = useMonthlyMetrics(entityKind, entityID);
  return (
    <>
      {(readQuota || writeQuota) && (
        <Grid container spacing={2} item xs={12}>
          {readQuota && <UsageIndicator standalone={true} kind="read" usage={latest.readBytes} quota={readQuota} />}
          {writeQuota && <UsageIndicator standalone={true} kind="write" usage={latest.writeBytes} quota={writeQuota} />}
        </Grid>
      )}
      <TopIndicators latest={latest} total={total} />
      {error && <ErrorNote error={error} />}
    </>
  );
};

export interface MetricsWeekProps {
  entityKind: EntityKind;
  entityID: string;
  field: string;
  title: string;
}

const MetricsWeek: FC<MetricsWeekProps> = ({ entityKind, entityID, field, title }) => {
  const { metrics, total, latest, error } = useWeeklyMetrics(entityKind, entityID);
  return (
    <>
      <WeekChart metrics={metrics} y1={field} title={title} />
      {error && <ErrorNote error={error} />}
    </>
  );
};