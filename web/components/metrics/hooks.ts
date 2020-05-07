import { useQuery } from "@apollo/react-hooks";

import { GET_METRICS } from "../../apollo/queries/metrics";
import { GetMetrics, GetMetricsVariables } from "../../apollo/types/GetMetrics";
import { EntityKind } from "../../apollo/types/globalTypes";
import { hourFloor, monthFloor, normalizeMetrics, now, weekAgo, yearAgo } from "./util";

export const useWeeklyMetrics = (entityKind: EntityKind, entityID: string) => {
  const from = hourFloor(weekAgo());
  const until = hourFloor(now());

  const { loading, error, data } = useQuery<GetMetrics, GetMetricsVariables>(GET_METRICS, {
    variables: {
      entityKind,
      entityID,
      from: from.toISOString(),
      period: "H",
    },
  });

  const { metrics, total, latest } = normalizeMetrics(from, until, "hour", data?.getMetrics);

  return { metrics, total, latest, error, loading };
};

export const useMonthlyMetrics = (entityKind: EntityKind, entityID: string) => {
  const from = monthFloor(yearAgo());
  const until = monthFloor(now());

  const { loading, error, data } = useQuery<GetMetrics, GetMetricsVariables>(GET_METRICS, {
    variables: {
      entityKind,
      entityID,
      from: from.toISOString(),
      period: "M",
    },
  });

  const { metrics, total, latest } = normalizeMetrics(from, until, "month", data?.getMetrics);

  return { metrics, total, latest, error, loading };
};
