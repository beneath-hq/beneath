export interface Metrics {
  readOps: number;
  readBytes: number;
  readRecords: number;
  writeOps: number;
  writeBytes: number;
  writeRecords: number;
}

export interface MetricsWithTime extends Metrics {
  time: Date | string;
}

export const blankMetrics = (): Metrics => {
  return {
    readOps: 0,
    readBytes: 0,
    readRecords: 0,
    writeOps: 0,
    writeBytes: 0,
    writeRecords: 0,
  };
};

export const aggregateMetrics = (metrics: Metrics[] | null): Metrics => {
  const result = blankMetrics();
  if (metrics) {
    for (const m of metrics) {
      result.readOps += m.readOps;
      result.readBytes += m.readBytes;
      result.readRecords += m.readRecords;
      result.writeOps += m.writeOps;
      result.writeBytes += m.writeBytes;
      result.writeRecords += m.writeRecords;
    }
  }
  return result;
};

export interface NormalizeMetricsResult {
  latest: Metrics;
  metrics: MetricsWithTime[];
  total: Metrics;
}

export const normalizeMetrics = (
  from: Date, until: Date, period: "hour" | "month", metrics: MetricsWithTime[] | null
): NormalizeMetricsResult => {
  // make sure metrics is well-formatted
  if ((metrics === null) || (metrics.length === 0)) {
    metrics = [{
      time: from,
      ...blankMetrics()
    }];
  }

  // add boundary time values if necessary
  const floorer = period === "hour" ? hourFloor : monthFloor;
  if (floorer(metrics[0].time).getTime() > floorer(from).getTime()) {
    metrics.splice(0, 0, { time: from, ...blankMetrics() });
  }
  if (floorer(metrics[metrics.length - 1].time).getTime() < floorer(until).getTime()) {
    metrics.push({ time: until, ...blankMetrics() });
  }

  // fill blanks
  const safeguard = 200;
  for (let i = 0; i < metrics.length - 1; i++) {
    const curr = floorer(metrics[i].time);
    const next = floorer(metrics[i + 1].time);

    if (next.getTime() <= curr.getTime()) {
      // shouldn't happen
      console.error("next metric isn't in the next period");
      continue;
    }

    const expected = new Date(curr);
    if (period === "hour") {
      expected.setUTCHours(expected.getUTCHours() + 1);
    } else {
      expected.setUTCMonth(expected.getUTCMonth() + 1);
    }

    if (next.getTime() !== expected.getTime()) {
      metrics.splice(i + 1, 0, { time: expected, ...blankMetrics() });
    }

    if (i > safeguard) {
      console.error("hit safeguard when normalizing metrics");
      break;
    }
  }

  // done
  return {
    latest: metrics[metrics.length - 1],
    metrics,
    total: aggregateMetrics(metrics),
  };
};

export const hourFloor = (date: Date | string): Date => {
  if (typeof date === "string") {
    date = new Date(date);
  }
  return new Date(Date.UTC(date.getUTCFullYear(), date.getUTCMonth(), date.getUTCDate(), date.getUTCHours(), 0, 0, 0));
};

export const monthFloor = (date: Date | string): Date => {
  if (typeof date === "string") {
    date = new Date(date);
  }
  return new Date(Date.UTC(date.getUTCFullYear(), date.getUTCMonth(), 1, 0, 0, 0, 0));
};

export const now = () => {
  return new Date();
};

export const weekAgo = () => {
  const now = new Date();
  const then = new Date(now.getTime() - 7 * 24 * 60 * 60 * 1000);
  return then;
};

export const yearAgo = () => {
  const now = new Date();
  const then = new Date(now.getTime() - 365 * 24 * 60 * 60 * 1000);
  return then;
};
