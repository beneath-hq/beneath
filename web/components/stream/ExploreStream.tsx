import { useRecords } from "beneath-react";
import React, { FC, useEffect, useState } from "react";

import { Box, makeStyles, Theme } from "@material-ui/core";
import Button from "@material-ui/core/Button";
import Chip from "@material-ui/core/Chip";
import Grid from "@material-ui/core/Grid";
import Typography from "@material-ui/core/Typography";
import AddBoxIcon from "@material-ui/icons/AddBox";
import FiberManualRecordIcon from "@material-ui/icons/FiberManualRecord";
import OpenInNewIcon from "@material-ui/icons/OpenInNew";
import ToggleButton from "@material-ui/lab/ToggleButton";
import ToggleButtonGroup from "@material-ui/lab/ToggleButtonGroup";

import { StreamByOrganizationProjectAndName_streamByOrganizationProjectAndName } from "../../apollo/types/StreamByOrganizationProjectAndName";
import { useToken } from "../../hooks/useToken";
import BNTextField from "../BNTextField";
import { Link } from "../Link";
import Loading from "../Loading";
import SelectField from "../SelectField";
import VSpace from "../VSpace";
import RecordsTable from "./RecordsTable";
import { Schema } from "./schema";

interface ExploreStreamProps {
  stream: StreamByOrganizationProjectAndName_streamByOrganizationProjectAndName;
  setLoading: (loading: boolean) => void;
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
    width: "4rem",
    height: "3rem",
    borderColor: "text.primary",
    borderRadius: "5%"
  },
}));

const ExploreStream: FC<ExploreStreamProps> = ({ stream, setLoading }: ExploreStreamProps) => {
  if (!stream.primaryStreamInstanceID) {
    throw Error("internal error: can't use ExploreStream for stream without a primary instance ID");
  }

  // determine if stream may have more data incoming
  const finalized = !!stream.primaryStreamInstance?.madeFinalOn;

  // state
  // const [queryType, setQueryType] = useState<"log" | "index">(finalized ? "index" : "log");
  const [queryType, setQueryType] = useState<"log" | "index">("index");
  const [logPeek, setLogPeek] = useState(finalized ? false : true);
  const [pendingFilter, setPendingFilter] = useState(""); // the value in the text field
  const [filter, setFilter] = useState(""); // updated when text field is submitted (used in call to useRecords)

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
      instanceID: stream.primaryStreamInstanceID,
    },
    query:
      queryType === "index"
        ? { type: "index", filter: filter === "" ? undefined : filter }
        : { type: "log", peek: logPeek },
    pageSize: 25,
    subscribe:
      typeof window === "undefined"
        ? false
        : finalized
        ? false
        : {
            pageSize: 100,
            pollFrequencyMs: 250,
          },
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
              <Button variant="outlined">
                <Typography>Write a record</Typography>
                <AddBoxIcon />
              </Button>
            </Grid>
            <Grid item>
              <Button variant="outlined">
                <Typography>Query in SQL editor</Typography>
                <OpenInNewIcon />
              </Button>
            </Grid>
          </Grid>
        </Grid>
        <Grid item>
          <form
            onSubmit={(e) => {
              e.preventDefault();
              if (pendingFilter !== filter) {
                setFilter(pendingFilter);
              }
            }}
          >
            <Grid container direction="column">
              <Grid item container spacing={2} justify="space-between">
                <Grid item>
                  <Grid container alignItems="center" spacing={2}>
                    <Grid item>
                      <ToggleButtonGroup
                        exclusive
                        size="small"
                        value={queryType}
                        onChange={(event, value) => setQueryType(value as "log" | "index")}
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
                      <Chip
                        label="Live"
                        variant="outlined"
                        color="primary"
                        size="small"
                        icon={<FiberManualRecordIcon />}
                      />
                    </Grid>
                  </Grid>
                </Grid>
                <Grid item>
                  {queryType === "log" && (
                    <>
                      <Grid container direction="row" alignItems="center" spacing={2}>
                        <Grid item>
                          <Typography>See the code</Typography>
                        </Grid>
                        <Grid item>
                          <ToggleButton value="true" size="small">
                            <Grid container direction="column">
                              <Grid item>
                                <Typography>Newest</Typography>
                              </Grid>
                              <Grid item>
                                <Typography>Oldest</Typography>
                              </Grid>
                            </Grid>
                          </ToggleButton>
                        </Grid>
                      </Grid>
                    </>
                    // OLD. TODO: reuse the setLogPeek()
                    // <Grid item sm={"auto"}>
                    //   <SelectField
                    //     id="peek"
                    //     label="Order"
                    //     value={logPeek ? "true" : "false"}
                    //     options={[
                    //       { label: "Latest first", value: "true" },
                    //       { label: "Oldest first", value: "false" },
                    //     ]}
                    //     onChange={({ target }) => setLogPeek(target.value === "true")}
                    //     controlClass={classes.selectPeekControl}
                    //   />
                    // </Grid>
                  )}
                </Grid>
              </Grid>
              {queryType === "index" && (
                <>
                  <Box border={1} className={classes.indexQueryBox} display="flex" alignItems="center" p={1}>
                    <Grid item container alignItems="center" spacing={2}>
                      <Grid item xs>
                        <Grid container alignItems="center" spacing={2}>
                          <Grid item>
                            <Box border={1} className={classes.indexQueryInput}>
                              {/* <BNTextField
                            id="where"
                            label="Filter"
                            value={pendingFilter}
                            margin="none"
                            onChange={({ target }) => setPendingFilter(target.value)}
                            fullWidth
                          /> */}
                            </Box>
                          </Grid>
                          <Grid item>
                            <Typography>
                              + Add condition
                            </Typography>
                          </Grid>
                        </Grid>
                      </Grid>
                      <Grid item>
                        <Typography>See the code</Typography>
                      </Grid>
                      <Grid item>
                        <Button
                          type="submit"
                          variant="contained"
                          color="primary"
                          disabled={
                            loading ||
                            !(pendingFilter.length === 0 || isJSON(pendingFilter)) ||
                            !(pendingFilter.length <= 1024)
                          }
                        >
                          Go
                        </Button>
                      </Grid>
                    </Grid>
                  </Box>
                </>
              )}
            </Grid>
          </form>
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

const isJSON = (val: string): boolean => {
  try {
    JSON.parse(val);
    return true;
  } catch {
    return false;
  }
};
