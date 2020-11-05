import { useQuery } from "@apollo/client";

import { GET_USAGE } from "../../apollo/queries/usage";
import { GetUsage, GetUsageVariables } from "../../apollo/types/GetUsage";
import { EntityKind } from "../../apollo/types/globalTypes";
import { hourFloor, monthFloor, normalizeMetrics, now, weekAgo, yearAgo } from "./util";

export const useWeeklyMetrics = (entityKind: EntityKind, entityID: string) => {
  const from = hourFloor(weekAgo());
  const until = hourFloor(now());

  const { loading, error, data } = useQuery<GetUsage, GetUsageVariables>(GET_USAGE, {
    variables: {
      entityKind,
      entityID,
      from: from.toISOString(),
      period: "H",
    },
  });

  const { metrics, total, latest } = normalizeMetrics(from, until, "hour", data?.getUsage);

  return { metrics, total, latest, error, loading };
};

export const useMonthlyMetrics = (entityKind: EntityKind, entityID: string) => {
  const from = monthFloor(yearAgo());
  const until = monthFloor(now());

  const { loading, error, data } = useQuery<GetUsage, GetUsageVariables>(GET_USAGE, {
    variables: {
      entityKind,
      entityID,
      from: from.toISOString(),
      period: "M",
    },
    fetchPolicy: "cache-and-network",
    pollInterval: 30000 // 30 seconds in milliseconds
  });

  const { metrics, total, latest } = normalizeMetrics(from, until, "month", data?.getUsage);

  return { metrics, total, latest, error, loading };
};
