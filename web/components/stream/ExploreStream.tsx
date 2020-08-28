import { useRecords } from "beneath-react";
import React, { FC, useEffect, useState } from "react";
import avro from "avsc";

import ArrowDownwardIcon from "@material-ui/icons/ArrowDownward";
import { Box, makeStyles, Theme, Dialog, DialogContent, DialogActions } from "@material-ui/core";
import Button from "@material-ui/core/Button";
import Chip from "@material-ui/core/Chip";
import Grid from "@material-ui/core/Grid";
import Typography from "@material-ui/core/Typography";
import AddBoxIcon from "@material-ui/icons/AddBox";
import FiberManualRecordIcon from "@material-ui/icons/FiberManualRecord";
import OpenInNewIcon from "@material-ui/icons/OpenInNew";
import ToggleButton from "@material-ui/lab/ToggleButton";
import ToggleButtonGroup from "@material-ui/lab/ToggleButtonGroup";

import { StreamByOrganizationProjectAndName_streamByOrganizationProjectAndName, StreamByOrganizationProjectAndName_streamByOrganizationProjectAndName_primaryStreamInstance } from "../../apollo/types/StreamByOrganizationProjectAndName";
import { useToken } from "../../hooks/useToken";
import Loading from "../Loading";
import VSpace from "../VSpace";
import RecordsTable from "./RecordsTable";
import { Schema } from "./schema";
import WriteStream from "../../components/stream/WriteStream";
import { StreamInstancesByOrganizationProjectAndStreamName_streamInstancesByOrganizationProjectAndStreamName } from "apollo/types/StreamInstancesByOrganizationProjectAndStreamName";
import CodeBlock from "components/CodeBlock";
import FilterForm from "./FilterForm";

interface ExploreStreamProps {
  stream: StreamByOrganizationProjectAndName_streamByOrganizationProjectAndName;
  instance: StreamInstancesByOrganizationProjectAndStreamName_streamInstancesByOrganizationProjectAndStreamName | StreamByOrganizationProjectAndName_streamByOrganizationProjectAndName_primaryStreamInstance | null;
  permissions: boolean;
  setLoading: (loading: boolean) => void;
}

// from schema.tsx
interface Column {
  name: string;
  type: avro.Type;
  actualType: avro.Type;
  doc?: string;
}

const useStyles = makeStyles((theme: Theme) => ({
  fetchMoreButton: {},
  infoCaption: {
    color: theme.palette.text.disabled,
  },
  errorCaption: {
    color: theme.palette.error.main,
  },
  selectQueryControl: {
    minWidth: 100,
  },
  selectPeekControl: {
    minWidth: 150,
  },
  indexQueryBox: {
    width: "100%",
    height: "4rem",
    borderColor: "text.primary",
  },
  indexQueryInput: {
    width: "25rem",
    height: "3rem",
    borderColor: "text.primary",
    borderRadius: "5%"
  },
}));

const ExploreStream: FC<ExploreStreamProps> = ({ stream, instance, permissions, setLoading }: ExploreStreamProps) => {
  if (!stream.primaryStreamInstanceID || !instance?.streamInstanceID) {
    return (
      <>
        <Typography>
          TODO: Create a mini tutorial for writing first data to the stream
        </Typography>
      </>
    );
  }

  // determine if stream may have more data incoming
  const finalized = !!stream.primaryStreamInstance?.madeFinalOn;

  // state
  const [queryType, setQueryType] = useState<"log" | "index">(finalized ? "index" : "log");
  const [logPeek, setLogPeek] = useState(finalized ? false : true);
  const [writeDialog, setWriteDialog] = React.useState(false);  // opens up the Write-a-Record dialog
  const [logCodeDialog, setLogCodeDialog] = React.useState(false); // opens up the See-the-Code dialog for the Log view
  const [indexCodeDialog, setIndexCodeDialog] = React.useState(false); // opens up the See-the-Code dialog for the Index view
  const [subscribeToggle, setSubscribeToggle] = React.useState(true); // updated by the LIVE/PAUSED toggle (used in call to useRecords)
  const [filter, setFilter] = React.useState(""); // used in call to useRecords

  // optimization: initializing a schema is expensive, so we keep it as state and reload it if stream changes
  const [schema, setSchema] = useState(() => new Schema(stream));
  useEffect(() => {
    if (schema.streamID !== stream.streamID) {
      setSchema(new Schema(stream));
    }
  }, [stream]);

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
    subscribe:
        isSubscribed(finalized, subscribeToggle)
        ? {
            pageSize: 100,
            pollFrequencyMs: 250,
          }
        : false,
    renderFrequencyMs: 250,
    maxRecords: 1000,
    flashDurationMs: 2000,
  });

  useEffect(() => {
    setLoading(subscription.online);
    return function cleanup() {
      setLoading(false);
    };
  }, [subscription.online]);

  const classes = useStyles();

  const Message: FC<{ children: string, error?: boolean }> = ({ error, children }) => (
    <Typography className={error ? classes.errorCaption : classes.infoCaption} variant="body2" align="center">
      {children}
    </Typography>
  );

  return (
    <>
      <Grid container direction="column" spacing={2}>
        <Grid item>
          <Grid container justify="flex-end" spacing={2}>
            <Grid item>
              {stream.allowManualWrites && (
                <>
                  <Button
                    variant="outlined"
                    onClick={() => {
                      setWriteDialog(true);
                    }}
                    endIcon={<AddBoxIcon/>}
                  >
                    Write a record
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
                      <WriteStream stream={stream} />
                    </DialogContent>
                  </Dialog>
                </>
              )}
            </Grid>
            <Grid item>
              <Button variant="outlined" endIcon={<OpenInNewIcon/>}>
                Query in SQL editor
              </Button>
            </Grid>
          </Grid>
        </Grid>
        <Grid item>
          <Grid container direction="column">
            <Grid item container spacing={2} justify="space-between">
              <Grid item>
                <Grid container alignItems="center" spacing={2}>
                  <Grid item>
                    <ToggleButtonGroup
                      exclusive
                      size="small"
                      value={queryType}
                      onChange={(_, value: "log" | "index" | null) => {
                        if (value !== null) setQueryType(value);
                      }}
                    >
                      <ToggleButton value="log">
                        <Grid container direction="column">
                          <Grid item>
                            <Typography>Log</Typography>
                          </Grid>
                          <Grid item>
                            <Typography>Sort by time written</Typography>
                          </Grid>
                        </Grid>
                      </ToggleButton>
                      <ToggleButton value="index">
                        <Grid container direction="column">
                          <Grid item>
                            <Typography>Index</Typography>
                          </Grid>
                          <Grid item>
                            <Typography>Lookup by key</Typography>
                          </Grid>
                        </Grid>
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
                        icon={<FiberManualRecordIcon style={{ color: "#6FCF97" }} />}
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
                        icon={<FiberManualRecordIcon style={{ color: "#8D919B" }} />}
                      />
                    )}
                  </Grid>
                </Grid>
              </Grid>
              <Grid item>
                {queryType === "log" && (
                  <>
                    <Grid container direction="row" alignItems="center" spacing={2}>
                      <Grid item>
                        <Button
                          onClick={() => {
                            setLogCodeDialog(true);
                          }}
                        >
                          See the code
                        </Button>
                        <Dialog open={logCodeDialog} onBackdropClick={() => setLogCodeDialog(false)}>
                          <DialogContent>
                            <CodeBlock language={"javascript"}>
                              {`import { useRecords } from "beneath-react";
const { records, error, loading, fetchMore, fetchMoreChanges, subscription, truncation } = useRecords({
  ${permissions ? "" : `secret: "YOUR_SECRET",\n  `}stream: "${stream.project.organization.name}/${
                                stream.project.name
                              }/${stream.name}",
  query: {type: "log", peek: ${logPeek}},
  pageSize: 25,
  subscribe: ${isSubscribed(finalized, subscribeToggle)},
  renderFrequencyMs: 250,
  maxRecords: 1000,
  flashDurationMs: 2000,
});`}
                            </CodeBlock>
                          </DialogContent>
                          <DialogActions>
                            <Button onClick={() => setLogCodeDialog(false)} color="primary">
                              Close
                            </Button>
                          </DialogActions>
                        </Dialog>
                      </Grid>
                      <Grid item>
                        <Button size="small" onClick={() => setLogPeek(!logPeek)}>
                          <ArrowDownwardIcon />
                          {logPeek && (
                            <Grid container direction="column">
                              <Grid item>
                                <Typography color="primary">Newest</Typography>
                              </Grid>
                              <Grid item>
                                <Typography>Oldest</Typography>
                              </Grid>
                            </Grid>
                          )}
                          {!logPeek && (
                            <Grid container direction="column">
                              <Grid item>
                                <Typography>Oldest</Typography>
                              </Grid>
                              <Grid item>
                                <Typography color="primary">Newest</Typography>
                              </Grid>
                            </Grid>
                          )}
                        </Button>
                      </Grid>
                    </Grid>
                  </>
                )}
              </Grid>
            </Grid>
            {queryType === "index" && (
              <>
                <Box border={1} className={classes.indexQueryBox} display="flex" alignItems="center" p={1}>
                  <Grid item container alignItems="center" spacing={2}>
                    <Grid item xs>
                      <FilterForm
                        index={schema.columns.filter((col) => col.key) as Column[]}
                        onChange={(filter) => setFilter(filter)}
                      />
                    </Grid>
                    <Grid item>
                      <Button
                        onClick={() => {
                          setIndexCodeDialog(true);
                        }}
                      >
                        See the code
                      </Button>
                      <Dialog open={indexCodeDialog} onBackdropClick={() => setIndexCodeDialog(false)}>
                        <DialogContent>
                          <CodeBlock language={"javascript"}>
                            {`import { useRecords } from "beneath-react";
const { records, error, loading, fetchMore, fetchMoreChanges, subscription, truncation } = useRecords({
  ${permissions ? "" : `secret: "YOUR_SECRET",\n  `}stream: "${stream.project.organization.name}/${
                              stream.project.name
                            }/${stream.name}",
  query: {type: "index", filter: ${filter === "" ? undefined : filter}},
  pageSize: 25,
  subscribe: ${isSubscribed(finalized, subscribeToggle)},
  renderFrequencyMs: 250,
  maxRecords: 1000,
  flashDurationMs: 2000,
});`}
                          </CodeBlock>
                        </DialogContent>
                        <DialogActions>
                          <Button onClick={() => setIndexCodeDialog(false)} color="primary">
                            Close
                          </Button>
                        </DialogActions>
                      </Dialog>
                    </Grid>
                  </Grid>
                </Box>
              </>
            )}
          </Grid>
          {filter !== "" && error && <Message error={true}>{error.message}</Message>}
          {truncation.start && <Message>You loaded so many more rows that we had to remove some from the top</Message>}
          {subscription.error && <Message error={true}>{subscription.error.message}</Message>}
          {loading && records.length === 0 && <Loading justify="center" />}
          {(!loading || records.length > 0) && (
            <RecordsTable schema={schema} records={records} showTimestamps={queryType === "log"} />
          )}
          <VSpace units={4} />
          {truncation.end && <Message>We removed some records from the bottom to fit new records in the table</Message>}
          {fetchMore && (
            <Grid container justify="center">
              <Grid item>
                <Button
                  variant="outlined"
                  color="primary"
                  className={classes.fetchMoreButton}
                  disabled={loading}
                  onClick={() => fetchMore()}
                >
                  Fetch more
                </Button>
              </Grid>
            </Grid>
          )}
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
          <VSpace units={8} />
        </Grid>
      </Grid>
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
