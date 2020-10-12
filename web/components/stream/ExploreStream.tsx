import {
  Button,
  Chip,
  Dialog,
  DialogActions,
  DialogContent,
  Grid,
  makeStyles,
  Theme,
  Typography,
} from "@material-ui/core";
import { Code } from "@material-ui/icons";
import ArrowDownwardIcon from "@material-ui/icons/ArrowDownward";
import AddBoxIcon from "@material-ui/icons/AddBox";
import FiberManualRecordIcon from "@material-ui/icons/FiberManualRecord";
import { ToggleButton, ToggleButtonGroup } from "@material-ui/lab";
import { useRecords } from "beneath-react";
import React, { FC, useEffect, useState } from "react";

import { StreamByOrganizationProjectAndName_streamByOrganizationProjectAndName} from "../../apollo/types/StreamByOrganizationProjectAndName";
import { useToken } from "../../hooks/useToken";
import Loading from "../Loading";
import RecordsTable from "./RecordsTable";
import { Schema } from "./schema";
import WriteStream from "../../components/stream/WriteStream";
import CodeBlock from "components/CodeBlock";
import FilterForm from "./FilterForm";
import { NakedLink } from "components/Link";
import VSpace from "components/VSpace";
import clsx from "clsx";
import { Instance } from "pages/stream";

interface ExploreStreamProps {
  stream: StreamByOrganizationProjectAndName_streamByOrganizationProjectAndName;
  instance: Instance | null;
}

const useStyles = makeStyles((theme: Theme) => ({
  topRowHeight: {
    height: 28,
  },
  toggleButton: {
    width: 100,
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
  justifyLeftXsSm: {
    [theme.breakpoints.down("sm")]: {
      justifyContent: "left"
    }
  },
  fetchMoreButton: {},
}));

const ExploreStream: FC<ExploreStreamProps> = ({ stream, instance }: ExploreStreamProps) => {
  if (!instance) {
    return (
      <>
        <Typography>TODO: Create a mini tutorial for writing first data to the stream</Typography>
      </>
    );
  }

  // determine if stream may have more data incoming
  const finalized = !!instance.madeFinalOn;
  const isPublic = stream.project.public;

  // state
  const [queryType, setQueryType] = useState<"log" | "index">(finalized ? "index" : "log");
  const [logPeek, setLogPeek] = useState(finalized ? false : true);
  const [writeDialog, setWriteDialog] = React.useState(false); // opens up the Write-a-Record dialog
  const [indexCodeDialog, setIndexCodeDialog] = React.useState(false); // opens up the See-the-Code dialog for the Index view
  const [subscribeToggle, setSubscribeToggle] = React.useState(true); // updated by the LIVE/PAUSED toggle (used in call to useRecords)
  const [filter, setFilter] = React.useState(""); // used in call to useRecords

  // optimization: initializing a schema is expensive, so we keep it as state and reload it if stream changes
  const [schema, setSchema] = useState(() => new Schema(stream.avroSchema, stream.streamIndexes));
  useEffect(() => {
    setSchema(new Schema(stream.avroSchema, stream.streamIndexes));
  }, [stream.streamID]);

  // get records
  const token = useToken();
  const { records, error, loading, fetchMore, fetchMoreChanges, subscription, truncation } = useRecords({
    secret: token || undefined,
    stream: {
      instanceID: instance?.streamInstanceID,
    },
    query:
      queryType === "index"
        ? { type: "index", filter: filter === "" ? undefined : filter }
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

  const Message: FC<{ children: string; error?: boolean }> = ({ error, children }) => (
    <Typography className={error ? classes.errorCaption : classes.infoCaption} variant="body2" align="center">
      {children}
    </Typography>
  );

  return (
    <>
      {/* top-row buttons */}
      <Grid container justify="space-between" alignItems="flex-start" spacing={2}>
        <Grid item xs={12} md={3}>
          <Grid container direction="column">
            <Grid item>
              <Grid container alignItems="flex-start" spacing={2}>
                <Grid item>
                  <ToggleButtonGroup
                    exclusive
                    size="small"
                    value={queryType}
                    onChange={(_, value: "log" | "index" | null) => {
                      if (value !== null) setQueryType(value);
                    }}
                  >
                    <ToggleButton value="log" className={clsx(classes.topRowHeight, classes.toggleButton)}>Log</ToggleButton>
                    <ToggleButton value="index" className={clsx(classes.topRowHeight, classes.toggleButton)}>Index</ToggleButton>
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
              </Grid>
            </Grid>
            <VSpace units={2} />
            {queryType === "log" && (
              <Grid item>
                <Typography variant="caption">Sort the stream by timestamp of each write.</Typography>
              </Grid>
            )}
            {queryType === "index" && (
              <Grid item>
                <Typography variant="caption">Query the stream by the key fields.</Typography>
              </Grid>
            )}
          </Grid>
        </Grid>
        <Grid item xs={12} md={6}>
          {queryType === "log" && (
            <>
              <Grid
                container
                direction="column"
                justify="center"
                alignItems="center"
                spacing={2}
                className={classes.justifyLeftXsSm}
              >
                <Grid item>
                  {logPeek && (
                    <Button onClick={() => setLogPeek(!logPeek)} size="small" startIcon={<ArrowDownwardIcon />} className={classes.topRowHeight}>
                      Newest to Oldest
                    </Button>
                  )}
                  {!logPeek && (
                    <Button onClick={() => setLogPeek(!logPeek)} size="small" startIcon={<ArrowDownwardIcon />} className={classes.topRowHeight}>
                      Oldest to Newest
                    </Button>
                  )}
                </Grid>
              </Grid>
            </>
          )}
          {queryType === "index" && (
            <>
              <Grid
                container
                direction="row"
                justify="center"
                alignItems="center"
                spacing={2}
                className={classes.justifyLeftXsSm}
              >
                <Grid item>
                  <FilterForm
                    index={schema.columns.filter((col) => col.isKey)}
                    onChange={(filter) => setFilter(filter)}
                  />
                </Grid>
                {filter && (
                  <Grid item>
                    <Button
                      onClick={() => {
                        setIndexCodeDialog(true);
                      }}
                      size="small"
                      endIcon={<Code />}
                      className={classes.topRowHeight}
                    >
                      Filter
                    </Button>
                    <Dialog open={indexCodeDialog} onBackdropClick={() => setIndexCodeDialog(false)}>
                      <DialogContent>
                        <CodeBlock language={"python"}>
                          {`${filter}`}
                        </CodeBlock>
                      </DialogContent>
                      <DialogActions>
                        <Button onClick={() => setIndexCodeDialog(false)} color="primary">
                          Close
                        </Button>
                      </DialogActions>
                    </Dialog>
                  </Grid>
                )}
              </Grid>
            </>
          )}
        </Grid>
        <Grid item xs={12} md={3}>
          <Grid container spacing={2} justify="flex-end" className={classes.justifyLeftXsSm}>
            {stream.allowManualWrites && (
              <>
                <Grid item>
                  <Button
                    variant="outlined"
                    onClick={() => {
                      setWriteDialog(true);
                    }}
                    endIcon={<AddBoxIcon />}
                    size="small"
                    className={classes.topRowHeight}
                  >
                    Write record
                  </Button>
                  <Dialog
                    open={writeDialog}
                    fullWidth={true}
                    maxWidth={"xs"}
                    onBackdropClick={() => {
                      setWriteDialog(false);
                    }}
                  >
                    {/*
                    // disable for now
                    // must update js client to be able to write data (current local resolvers do not work anymore!)
                    // and to allow both stream and batch writes
                    */}
                    <DialogContent>
                      <WriteStream
                        stream={stream}
                        instanceID={instance.streamInstanceID}
                        setWriteDialog={setWriteDialog}
                      />
                    </DialogContent>
                  </Dialog>
                </Grid>
              </>
            )}
            <Grid item>
              <Button
                variant="outlined"
                component={NakedLink}
                href={`/-/sql?stream=${stream.project.organization.name}/${stream.project.name}/${stream.name}`}
                as={`/-/sql`}
                size="small"
                className={classes.topRowHeight}
              >
                Query with SQL
              </Button>
            </Grid>
          </Grid>
        </Grid>
      </Grid>
      {/* records table */}
      <VSpace units={2} />
      {filter !== "" && error && <Message error={true}>{error.message}</Message>}
      {truncation.start && <Message>You loaded so many more rows that we had to remove some from the top</Message>}
      {subscription.error && <Message error={true}>{subscription.error.message}</Message>}
      {loading && records.length === 0 && <Loading justify="center" />}
      {(!loading || records.length > 0) && (
        <RecordsTable
          paper
          schema={schema}
          records={records}
          fetchMore={fetchMore}
          showTimestamps={queryType === "log"}
          error={undefined} /* Todo */
          callToAction={undefined} /* Todo */
        />
      )}
      {truncation.end && <Message>We removed some records from the bottom to fit new records in the table</Message>}
      {!fetchMore && fetchMoreChanges && (
        <Grid container justify="center">
          <Grid item>
            <Button
              variant="outlined"
              color="primary"
              className={classes.fetchMoreButton}
              disabled={loading}
              onClick={() => fetchMoreChanges()}
            >
              Fetch more changes
            </Button>
          </Grid>
        </Grid>
      )}
      {filter === "" && error && <Message error={true}>{error.message}</Message>}
      {!loading && !fetchMore && !fetchMoreChanges && !truncation.start && !truncation.end && (
        <Message>
          {`${records.length === 0 ? "Found no rows" : "Loaded all rows"} ${
            filter !== "" ? "that match the filter" : ""
          }`}
        </Message>
      )}
    </>
  );
};

export default ExploreStream;

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
