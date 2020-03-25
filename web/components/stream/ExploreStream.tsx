import { FetchMoreFunction, useRecords } from "beneath-react";

import React, { FC, useEffect, useState } from "react";

import { makeStyles, Theme } from "@material-ui/core";
import Button from "@material-ui/core/Button";
import Grid from "@material-ui/core/Grid";
import Typography from "@material-ui/core/Typography";

import { QueryStream_streamByProjectAndName } from "../../apollo/types/QueryStream";
import { useToken } from "../../hooks/useToken";
import BNTextField from "../BNTextField";
import LinkTypography from "../LinkTypography";
import Loading from "../Loading";
import SelectField from "../SelectField";
import VSpace from "../VSpace";
import RecordsTable from "./RecordsTable";
import { Schema } from "./schema";

interface ExploreStreamProps {
  stream: QueryStream_streamByProjectAndName;
  setLoading: (loading: boolean) => void;
}

const useStyles = makeStyles((theme: Theme) => ({
  submitButton: {
    marginTop: theme.spacing(1.5),
  },
  fetchMoreButton: {},
  noMoreDataCaption: {
    color: theme.palette.text.disabled,
  },
  errorCaption: {
    color: theme.palette.error.main,
  },
  selectViewControl: {
    minWidth: 100,
  },
}));

const ExploreStream: FC<ExploreStreamProps> = ({ stream, setLoading }: ExploreStreamProps) => {
  // state
  const [view, setView] = useState<"lookup" | "log" | "latest">("lookup");
  const [pageSize, setPageSize] = useState(25);
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
  const { records, error, loading, subscribed, fetchMore } = useRecords({
    secret: token || undefined,
    project: stream.project.name,
    stream: stream.name,
    instanceID: stream.currentStreamInstanceID || undefined,
    view,
    pageSize,
    filter: filter === "" ? undefined : filter,
    subscribe: false,
  });

  const classes = useStyles();
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
          <Grid item xs={12} sm={"auto"}>
            <SelectField
              id="view"
              label="View"
              value={view}
              options={[
                { label: "Lookup", value: "lookup" },
                { label: "Latest", value: "latest" },
                { label: "Log", value: "log" },
              ]}
              onChange={({ target }) => setView(target.value as "lookup" | "log" | "latest")}
              controlClass={classes.selectViewControl}
            />
          </Grid>
          {view === "lookup" && (
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
                  errorText={filter !== "" && error ? error.message : undefined}
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
      <VSpace units={2} />
      {loading && records.length === 0 && <Loading justify="center" />}
      {(!loading || records.length > 0) && (
        <RecordsTable schema={schema} records={records} showTimestamps={view !== "lookup"} />
      )}
      <VSpace units={4} />
      {fetchMore && (
        <Grid container justify="center">
          <Grid item>
            <Button
              variant="outlined"
              color="primary"
              className={classes.fetchMoreButton}
              disabled={loading}
              onClick={() => fetchMore({ pageSize })}
            >
              Fetch more
            </Button>
          </Grid>
        </Grid>
      )}
      {!loading && !fetchMore && (
        <Typography className={classes.noMoreDataCaption} variant="body2" align="center">
          {records.length === 0 ? "Found no rows" : "Loaded all rows"}
          {" "}
          {filter !== "" ? "that match the filter" : ""}
        </Typography>
      )}
      {error && (
        <Typography className={classes.errorCaption} variant="body2" align="center">
          {error.message}
        </Typography>
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
