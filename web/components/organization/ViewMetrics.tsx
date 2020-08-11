import { Grid } from "@material-ui/core";
import { FC } from "react";

import { EntityKind } from "../../apollo/types/globalTypes";
import { OrganizationByName_organizationByName_PrivateOrganization } from "../../apollo/types/OrganizationByName";
import ErrorNote from "../ErrorNote";
import { useMonthlyMetrics, useWeeklyMetrics } from "../metrics/hooks";
import TopIndicators from "../metrics/user/TopIndicators";
import UsageIndicator from "../metrics/user/UsageIndicator";
import WeekChart from "../metrics/WeekChart";

export interface ViewMetricsProps {
  organization: OrganizationByName_organizationByName_PrivateOrganization;
}

const ViewMetrics: FC<ViewMetricsProps> = ({ organization }) => {
  // show metrics on org level, unless it's a personal org that doesn't handle billing (then show on user-level)
  let props: MetricsProps;
  if (organization.personalUser && organization.personalUser.billingOrganizationID !== organization.organizationID) {
    props = {
      entityKind: EntityKind.User,
      entityID: organization.personalUser.userID,
      readQuota: organization.personalUser.readQuota,
      writeQuota: organization.personalUser.writeQuota,
      scanQuota: organization.personalUser.scanQuota,
    };
  } else {
    props = {
      entityKind: EntityKind.Organization,
      entityID: organization.organizationID,
      readQuota: organization.prepaidReadQuota || organization.readQuota,
      writeQuota: organization.prepaidWriteQuota || organization.writeQuota,
      scanQuota: organization.prepaidScanQuota || organization.scanQuota,
    };
  }

  return (
    <Grid container spacing={2}>
      <MetricsOverview {...props} />
      <MetricsWeek {...props} />
    </Grid>
  );
};

export default ViewMetrics;

export interface MetricsProps {
  entityKind: EntityKind;
  entityID: string;
  readQuota?: number | null;
  writeQuota?: number | null;
  scanQuota?: number | null;
}

const MetricsOverview: FC<MetricsProps> = ({ entityKind, entityID, readQuota, writeQuota, scanQuota }) => {
  const { metrics, total, latest, error } = useMonthlyMetrics(entityKind, entityID);
  return (
    <>
      {(readQuota || writeQuota || scanQuota) && (
        <Grid container spacing={2} item xs={12}>
          {readQuota && <UsageIndicator standalone={true} kind="read" usage={latest.readBytes} quota={readQuota} />}
          {writeQuota && <UsageIndicator standalone={true} kind="write" usage={latest.writeBytes} quota={writeQuota} />}
          {scanQuota && <UsageIndicator standalone={true} kind="scan" usage={latest.scanBytes} quota={scanQuota} />}
        </Grid>
      )}
      <TopIndicators latest={latest} total={total} />
      {error && <ErrorNote error={error} />}
    </>
  );
};

const MetricsWeek: FC<MetricsProps> = ({ entityKind, entityID }) => {
  const { metrics, total, latest, error } = useWeeklyMetrics(entityKind, entityID);
  return (
    <>
      <WeekChart metrics={metrics} y1={"readBytes"} title="Data read in the last 7 days" />
      {error && <ErrorNote error={error} />}
    </>
  );
};
