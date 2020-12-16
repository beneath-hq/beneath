import { useQuery } from "@apollo/client";

import { GET_USAGE } from "apollo/queries/usage";
import { GetUsage, GetUsageVariables } from "apollo/types/GetUsage";
import { EntityKind, UsageLabel } from "apollo/types/globalTypes";

export const useHourlyUsage = (entityKind: EntityKind, entityID: string) => {
  const now = new Date();
  const weekAgo = new Date(now.getTime() - 7 * 24 * 60 * 60 * 1000);
  const from = hourFloor(weekAgo);
  const until = hourFloor(now);

  const { loading, error, data } = useQuery<GetUsage, GetUsageVariables>(GET_USAGE, {
    variables: {
      input: {
        entityKind: entityKind,
        entityID: entityID,
        label: UsageLabel.Hourly,
        from: from.toISOString(),
        until: until.toISOString(),
      },
    },
  });

  let usages: Usage[] | undefined;
  if (data) {
    usages = imputeHourlyUsages(from, until, data.getUsage);
  }

  return { data: usages, loading, error };
};

export const useQuotaUsage = (entityKind: EntityKind, entityID: string, quotaStartTime: string) => {
  const { loading, error, data } = useQuery<GetUsage, GetUsageVariables>(GET_USAGE, {
    variables: {
      input: {
        entityKind: entityKind,
        entityID: entityID,
        label: UsageLabel.QuotaMonth,
        from: quotaStartTime,
      },
    },
    fetchPolicy: "cache-and-network",
    pollInterval: 30000, // 30 seconds
  });

  if ((!data && loading) || error) {
    return { loading, error };
  }

  let usage = blankUsage();
  if (data?.getUsage.length) {
    usage = data?.getUsage[0];
  }

  return { data: usage, loading, error };
};

export const useTotalUsage = (entityKind: EntityKind, entityID: string) => {
  const { loading, error, data } = useQuery<GetUsage, GetUsageVariables>(GET_USAGE, {
    variables: {
      input: {
        entityKind: entityKind,
        entityID: entityID,
        label: UsageLabel.Total,
      },
    },
    fetchPolicy: "cache-and-network",
    pollInterval: 30000, // 30 seconds
  });

  if ((!data && loading) || error) {
    return { loading, error };
  }

  let usage = blankUsage();
  if (data?.getUsage.length) {
    usage = data?.getUsage[0];
  }

  return { data: usage, loading, error };
};

export interface Usage {
  time: ControlTime;
  readOps: number;
  readBytes: number;
  readRecords: number;
  writeOps: number;
  writeBytes: number;
  writeRecords: number;
  scanOps: number;
  scanBytes: number;
}

export const blankUsage = (): Usage => {
  return {
    time: "",
    readOps: 0,
    readBytes: 0,
    readRecords: 0,
    writeOps: 0,
    writeBytes: 0,
    writeRecords: 0,
    scanOps: 0,
    scanBytes: 0,
  };
};

export const imputeHourlyUsages = (from: Date, until: Date, metrics: Usage[]): Usage[] => {
  // make sure metrics is well-formatted
  if ((metrics === undefined) || (metrics.length === 0)) {
    metrics = [{ ...blankUsage(), time: from.toISOString() }];
  } else {
    // Apollo result arrays are immutable, so we clone to a new result array
    const res: Usage[] = [];
    for (const usage of metrics) {
      res.push({ ...usage });
    }
    metrics = res;
  }

  // add boundary time values if necessary
  if (hourFloor(metrics[0].time).getTime() > hourFloor(from).getTime()) {
    metrics.splice(0, 0, { ...blankUsage(), time: from.toISOString() });
  }
  if (hourFloor(metrics[metrics.length - 1].time).getTime() < hourFloor(until).getTime()) {
    metrics.push({ ...blankUsage(), time: until.toISOString() });
  }

  // fill blanks
  const safeguard = 200;
  for (let i = 0; i < metrics.length - 1; i++) {
    const curr = hourFloor(metrics[i].time);
    const next = hourFloor(metrics[i + 1].time);

    if (next.getTime() <= curr.getTime()) {
      // shouldn't happen
      console.error("next metric isn't in the next period");
      continue;
    }

    const expected = new Date(curr);
    expected.setUTCHours(expected.getUTCHours() + 1)

    if (next.getTime() !== expected.getTime()) {
      metrics.splice(i + 1, 0, { ...blankUsage(), time: expected.toISOString() });
    }

    if (i > safeguard) {
      console.error("hit safeguard when normalizing metrics");
      break;
    }
  }

  // done
  return metrics;
};

export const hourFloor = (date: Date | string): Date => {
  if (typeof date === "string") {
    date = new Date(date);
  }
  return new Date(Date.UTC(date.getUTCFullYear(), date.getUTCMonth(), date.getUTCDate(), date.getUTCHours(), 0, 0, 0));
};
