import { Grid, Typography } from "@material-ui/core";
import React, { FC } from "react";
import Moment from "react-moment";

import { Tile, TileProps } from "./Tile";
import { PaperGrid } from "components/Paper";
import { QuotaUsageIndicator } from "components/usage/UsageIndicator";
import useMe from "hooks/useMe";

export const MyUsageTile: FC<TileProps> = (tileProps) => {
  const me = useMe();

  if (!me || me.organizationID !== me.personalUser?.billingOrganizationID) {
    return <></>;
  }

  return (
    <Tile
      nopaper
      href={`/organization?organization_name=${me.name}&tab=monitoring`}
      as={`/${me.name}/-/monitoring`}
      {...tileProps}
    >
      <Typography variant="h3">Current usage</Typography>
      <Typography variant="subtitle2" color="textSecondary" gutterBottom>
        Period from <Moment date={me.quotaStartTime} format="D MMM" /> to <Moment date={me.quotaEndTime} format="D MMM" />
      </Typography>
      <PaperGrid variant="outlined" container spacing={2}>
        <Grid item xs={12}>
          <QuotaUsageIndicator
            label="Reads"
            dense
            usage={me.readUsage}
            quota={me.readQuota}
            prepaidQuota={me.prepaidReadQuota}
          />
        </Grid>
        <Grid item xs={12}>
          <QuotaUsageIndicator
            label="Writes"
            dense
            usage={me.writeUsage}
            quota={me.writeQuota}
            prepaidQuota={me.prepaidWriteQuota}
          />
        </Grid>
        <Grid item xs={12}>
          <QuotaUsageIndicator
            label="Scans"
            dense
            usage={me.scanUsage}
            quota={me.scanQuota}
            prepaidQuota={me.prepaidScanQuota}
          />
        </Grid>
      </PaperGrid>
    </Tile>
  );
};

export default MyUsageTile;
