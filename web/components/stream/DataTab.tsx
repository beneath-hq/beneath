import { useRecords } from "beneath-react";
import _ from "lodash";
import { Button, Chip, Grid, makeStyles, Theme, Tooltip, Typography } from "@material-ui/core";
import ArrowDownwardIcon from "@material-ui/icons/ArrowDownward";
import FiberManualRecordIcon from "@material-ui/icons/FiberManualRecord";
import { ToggleButton, ToggleButtonGroup } from "@material-ui/lab";
import { useRouter } from "next/router";
import React, { FC, useEffect, useState } from "react";

import { useToken } from "../../hooks/useToken";
import RecordsTable from "./RecordsTable";
import { Schema } from "./schema";
import WriteStream from "./WriteStream";
import FilterForm from "./FilterForm";
import { NakedLink } from "components/Link";
import VSpace from "components/VSpace";
import clsx from "clsx";
import { StreamInstance } from "components/stream/types";
import ContentContainer, { CallToAction } from "components/ContentContainer";
import { StreamInstanceByOrganizationProjectStreamAndVersion_streamInstanceByOrganizationProjectStreamAndVersion_stream } from "apollo/types/StreamInstanceByOrganizationProjectStreamAndVersion";
import { makeStreamAs, makeStreamHref } from "./urls";

interface DataTabProps {
  stream: StreamInstanceByOrganizationProjectStreamAndVersion_streamInstanceByOrganizationProjectStreamAndVersion_stream;
  instance: StreamInstance | null;
}

const useStyles = makeStyles((theme: Theme) => ({
  topRowHeight: {
    height: 28,
  },
  toggleButton: {
    width: 100,
  },
  toggleButtonLabel: {
    display: "block",
    width: "100%",
  },
  liveIcon: {
    color: theme.palette.success.light,
  },
  pausedIcon: {
    color: theme.palette.grey[500],
  },
  infoCaption: {
    color: theme.palette.text.disabled,
  },
  errorCaption: {
    color: theme.palette.error.main,
  },
  safariButtonFix: {
    whiteSpace: "nowrap",
  },
}));

const DataTab: FC<DataTabProps> = ({ stream, instance }) => {
  if (!instance) {
    const cta: CallToAction = {
      message: `Please select a stream version`,
    };
    return (
      <>
        <ContentContainer callToAction={cta} />
      </>
    );
  }

  const NO_FILTER = {};

  // determine if stream may have more data incoming
  const finalized = !!instance.madeFinalOn;

  // state
  const [logPeek, setLogPeek] = useState(finalized ? false : true);
  const [subscribeToggle, setSubscribeToggle] = React.useState(true); // updated by the LIVE/PAUSED toggle (used in call to useRecords)

  // optimization: initializing a schema is expensive, so we keep it as state and reload it if stream changes
  const [schema, setSchema] = useState(() => new Schema(stream.avroSchema, stream.streamIndexes));
  useEffect(() => {
    setSchema(new Schema(stream.avroSchema, stream.streamIndexes));
  }, [stream.streamID]);

  const router = useRouter();
  const [filter, setFilter] = useState<any>(() => {
    // checks to see if a filter was provided in the URL
    if (typeof router.query.filter !== "string") return NO_FILTER;

    // attempts to parse JSON
    let filter: any;
    try {
      filter = JSON.parse(router.query.filter);
    } catch {
      return NO_FILTER;
    }

    // checks that the filter's keys are in the stream's index
    const keys = Object.keys(filter);
    const index = schema.columns.filter((col) => col.isKey);
    for (const key of keys) {
      const col = index.find((col) => col.name === key);
      if (typeof col === "undefined") return NO_FILTER;
    }

    // if query submitted in form {"key": "value"}, convert it to form {"key": {"_eq": "value"}}
    for (const key of keys) {
      const val = filter[key];
      if (typeof val !== "object") {
        filter[key] = { _eq: val };
      }
    }

    return filter;
  });
  const [queryType, setQueryType] = useState<"log" | "index">(finalized || !_.isEmpty(filter) ? "index" : "log");

  // get records
  const token = useToken();
  const { records, error, loading, fetchMore, fetchMoreChanges, subscription, truncation } = useRecords({
    secret: token || undefined,
    stream: {
      instanceID: instance?.streamInstanceID,
    },
    query:
      queryType === "index"
        ? { type: "index", filter: _.isEmpty(filter) ? undefined : JSON.stringify(filter) }
        : { type: "log", peek: logPeek },
    pageSize: 25,
    subscribe: isSubscribed(finalized, subscribeToggle)
      ? {
          pageSize: 100,
          pollFrequencyMs: 250,
        }
      : false,
    renderFrequencyMs: 250,
    maxRecords: 1000,
    flashDurationMs: 2000,
  });

  const classes = useStyles();

  // set filter in URL
  useEffect(() => {
    let href = makeStreamHref(stream, instance);
    let as = makeStreamAs(stream, instance);
    if (queryType === "index") {
      const filterJSON = JSON.stringify(filter);
      const noFilterJSON = JSON.stringify(NO_FILTER);
      href = href + (filterJSON !== noFilterJSON ? `&filter=${encodeURIComponent(filterJSON)}` : "");
      as = as + (filterJSON !== noFilterJSON ? `?filter=${encodeURIComponent(filterJSON)}` : "");
    }
    router.replace(href, as);
  }, [JSON.stringify(filter), queryType]);

  // CTAs
  let containerCta: CallToAction | undefined;
  if (!fetchMore && fetchMoreChanges) {
    containerCta = {
      buttons: [{ label: "Fetch more changes", onClick: () => fetchMoreChanges() }],
    };
  }
  let tableCta: CallToAction | undefined;
  if (!loading && _.isEmpty(filter) && records.length === 0) {
    tableCta = {
      message: `There's no data in this stream instance`,
      buttons: [
        {
          label: "Go to API docs",
          href: makeStreamHref(stream, instance, "api"),
          as: makeStreamAs(stream, instance, "api"),
        },
      ],
    };
  }
  if (!loading && queryType === "index" && !_.isEmpty(filter) && records.length === 0) {
    tableCta = {
      message: `Found no rows that match the filter`,
    };
  }

  // NOTES
  let note: string | undefined;
  if (truncation.end) {
    note = "We removed some records from the bottom to fit new records in the table";
  }

  // Messages at the top of the table use this component
  const Message: FC<{ children: string; error?: boolean }> = ({ error, children }) => (
    <Typography className={error ? classes.errorCaption : classes.infoCaption} variant="body2" align="center">
      {children}
    </Typography>
  );

  return (
    <>
      <ContentContainer callToAction={containerCta}>
        {/* top-row buttons */}
        <Grid container spacing={1} alignItems="center">
          <Grid item>
            <ToggleButtonGroup
              exclusive
              size="small"
              value={queryType}
              onChange={(_, value: "log" | "index" | null) => {
                if (value !== null) setQueryType(value);
              }}
            >
              <ToggleButton value="log" className={clsx(classes.topRowHeight, classes.toggleButton)}>
                <Tooltip title="Sort the stream by the time of each write">
                  <span className={classes.toggleButtonLabel}>Log</span>
                </Tooltip>
              </ToggleButton>
              <ToggleButton value="index" className={clsx(classes.topRowHeight, classes.toggleButton)}>
                <Tooltip title="Query the stream by the key fields">
                  <span className={classes.toggleButtonLabel}>Index</span>
                </Tooltip>
              </ToggleButton>
            </ToggleButtonGroup>
          </Grid>
          <Grid item>
            {isSubscribed(finalized, subscribeToggle) && (
              <Chip
                label="Live"
                variant="outlined"
                size="small"
                clickable
                onClick={() => setSubscribeToggle(false)}
                icon={<FiberManualRecordIcon className={classes.liveIcon} />}
                className={classes.topRowHeight}
              />
            )}
            {!isSubscribed(finalized, subscribeToggle) && (
              <Chip
                label="Paused"
                variant="outlined"
                size="small"
                clickable={isSubscribeable(finalized) ? true : false}
                onClick={() => {
                  if (isSubscribeable(finalized)) {
                    setSubscribeToggle(true);
                  }
                  return;
                }}
                icon={<FiberManualRecordIcon className={classes.pausedIcon} />}
                className={classes.topRowHeight}
              />
            )}
          </Grid>
          {queryType === "log" && (
            <Grid item>
              {logPeek && (
                <Button
                  variant="outlined"
                  onClick={() => setLogPeek(!logPeek)}
                  size="small"
                  startIcon={<ArrowDownwardIcon />}
                  className={classes.topRowHeight}
                >
                  Newest to oldest
                </Button>
              )}
              {!logPeek && (
                <Button
                  variant="outlined"
                  onClick={() => setLogPeek(!logPeek)}
                  size="small"
                  startIcon={<ArrowDownwardIcon />}
                  className={classes.topRowHeight}
                >
                  Oldest to newest
                </Button>
              )}
            </Grid>
          )}
          {queryType === "index" && (
            <FilterForm
              filter={filter}
              index={schema.columns.filter((col) => col.isKey)}
              onChange={(filter: any) => setFilter({ ...filter })}
            />
          )}
          <Grid item xs>
            <Grid container spacing={1} justify="flex-end" wrap="nowrap">
              <Grid item>
                <WriteStream
                  stream={stream}
                  instanceID={instance.streamInstanceID}
                  buttonClassName={clsx(classes.topRowHeight, classes.safariButtonFix)}
                />
              </Grid>
              <Grid item>
                <Button
                  variant="outlined"
                  component={NakedLink}
                  href={`/-/sql?stream=${stream.project.organization.name}/${stream.project.name}/${stream.name}`}
                  as={`/-/sql`}
                  size="small"
                  classes={{ root: clsx(classes.topRowHeight, classes.safariButtonFix) }}
                  disabled={!stream.useWarehouse}
                >
                  Query with SQL
                </Button>
              </Grid>
            </Grid>
          </Grid>
        </Grid>
        {/* records table */}
        <VSpace units={3} />
        {truncation.start && <Message>You loaded so many more rows that we had to remove some from the top</Message>}
        {subscription.error && <Message error={true}>{subscription.error.message}</Message>}
        <RecordsTable
          paper
          schema={schema}
          records={records}
          fetchMore={fetchMore}
          showTimestamps={queryType === "log"}
          callToAction={tableCta}
          error={error?.message}
          note={note}
          loading={loading}
        />
      </ContentContainer>
    </>
  );
};

export default DataTab;

const isSubscribeable = (finalized: boolean) => {
  if (typeof window === "undefined" || finalized) {
    return false;
  } else {
    return true;
  }
};

const isSubscribed = (finalized: boolean, subscribeToggle: boolean) => {
  return isSubscribeable(finalized) ? subscribeToggle : false;
};
