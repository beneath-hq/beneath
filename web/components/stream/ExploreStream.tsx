import { useRecords } from "beneath-react";
import React, { FC, useEffect, useState } from "react";

import { makeStyles, Theme } from "@material-ui/core";
import Button from "@material-ui/core/Button";
import Grid from "@material-ui/core/Grid";
import Typography from "@material-ui/core/Typography";

import { StreamByOrganizationProjectAndName_streamByOrganizationProjectAndName } from "../../apollo/types/StreamByOrganizationProjectAndName";
import { useToken } from "../../hooks/useToken";
import BNTextField from "../BNTextField";
import LinkTypography from "../LinkTypography";
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
  submitButton: {
    marginTop: theme.spacing(1.5),
  },
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
}));

const ExploreStream: FC<ExploreStreamProps> = ({ stream, setLoading }: ExploreStreamProps) => {
  if (!stream.currentStreamInstanceID) {
    throw Error("internal error: can't use ExploreStream for stream without a current instance ID");
  }

  // state
  const [queryType, setQueryType] = useState<"log" | "index">("log");
  const [logPeek, setLogPeek] = useState(true);
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
      instanceID: stream.currentStreamInstanceID,
    },
    query:
      queryType === "index"
        ? { type: "index", filter: filter === "" ? undefined : filter }
        : { type: "log", peek: logPeek },
    pageSize: 25,
    subscribe:
      typeof window === "undefined"
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
      <form
        onSubmit={(e) => {
          e.preventDefault();
          if (pendingFilter !== filter) {
            setFilter(pendingFilter);
          }
        }}
      >
        <Grid container spacing={2}>
          <Grid item xs={queryType === "log" ? "auto" : 12} sm={"auto"}>
            <SelectField
              id="query"
              label="Query"
              value={queryType}
              options={[
                { label: "Log", value: "log" },
                { label: "Index", value: "index" },
              ]}
              onChange={({ target }) => setQueryType(target.value as "log" | "index")}
              controlClass={classes.selectQueryControl}
            />
          </Grid>
          {queryType === "log" && (
            <Grid item sm={"auto"}>
              <SelectField
                id="peek"
                label="Order"
                value={logPeek ? "true" : "false"}
                options={[
                  { label: "Latest first", value: "true" },
                  { label: "Oldest first", value: "false" },
                ]}
                onChange={({ target }) => setLogPeek(target.value === "true")}
                controlClass={classes.selectPeekControl}
              />
            </Grid>
          )}
          {queryType === "index" && (
            <>
              <Grid item xs>
                <BNTextField
                  id="where"
                  label="Filter"
                  value={pendingFilter}
                  margin="none"
                  onChange={({ target }) => setPendingFilter(target.value)}
                  helperText={
                    <>
                      You can query the stream on indexed fields, check out the{" "}
                      {/* tslint:disable-next-line: max-line-length */}
                      <LinkTypography href="https://about.beneath.dev/docs/using-the-explore-tab/">
                        docs
                      </LinkTypography>{" "}
                      for more info.
                    </>
                  }
                  fullWidth
                />
              </Grid>
              <Grid item>
                <Button
                  type="submit"
                  variant="outlined"
                  color="primary"
                  className={classes.submitButton}
                  disabled={
                    loading || !(pendingFilter.length === 0 || isJSON(pendingFilter)) || !(pendingFilter.length <= 1024)
                  }
                >
                  Execute
                </Button>
              </Grid>
            </>
          )}
        </Grid>
      </form>
      {filter !== "" && error && <Message error={true}>{error.message}</Message>}
      {truncation.start && <Message>You loaded so many more rows that we had to remove some from the top</Message>}
      {subscription.error && <Message error={true}>{subscription.error.message}</Message>}
      <VSpace units={2} />
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
