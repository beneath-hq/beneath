import { useApolloClient, ApolloQueryResult } from "@apollo/client";
import { Button, CircularProgress, Grid, makeStyles, Theme, Chip } from "@material-ui/core";
import { useWarehouse } from "beneath-react";
import _ from "lodash";
import { useRouter } from "next/router";
import numbro from "numbro";
import React, { useState, useMemo, useEffect } from "react";

import CodeEditor from "components/CodeEditor";
import RecordsTable from "components/stream/RecordsTable";
import { Schema } from "components/stream/schema";
import { useToken } from "hooks/useToken";
import { QUERY_STREAM } from "apollo/queries/stream";
import { StreamByOrganizationProjectAndName, StreamByOrganizationProjectAndNameVariables } from "apollo/types/StreamByOrganizationProjectAndName";
import { toBackendName } from "lib/names";
import StreamPreview from "./StreamPreview";

const useStyles = makeStyles((_: Theme) => ({
  statusAction: {
    display: "flex",
    alignItems: "center",
  },
}));

const Main = () => {
  // Prepopulate query text if &stream=... url param is set
  const router = useRouter();
  const [queryText, setQueryText] = useState(() => {
    const prepopulateStream = router.query.stream;
    if (prepopulateStream) {
      return `select count(*) from \`${prepopulateStream}\``;
    }
    return "";
  });

  // Running query
  const token = useToken();
  const {
    analyzeQuery,
    runQuery,
    loading,
    error,
    job,
    records,
    fetchMore,
  } = useWarehouse({ secret: token || undefined });

  // Compute Schema object for job.resultAvroSchema
  const schema = useMemo(() => {
    const schema = job?.resultAvroSchema ? new Schema(job?.resultAvroSchema, []) : undefined;
    return schema;
  }, [job?.resultAvroSchema]);

  // Extract stream paths from query
  const streamPaths = useMemo(() => {
    const matches = queryText.match(/\`[_\-a-z0-9]+\/[_\-a-z0-9]+\/[_\-a-z0-9]+\`/g);
    const paths = _.uniq(matches);
    return paths;
  }, [queryText]);

  // Lookup every stream path in query
  const client = useApolloClient();
  const [streams, setStreams] = useState<ApolloQueryResult<StreamByOrganizationProjectAndName>[]>([]);
  useEffect(() => {
    const results: ApolloQueryResult<StreamByOrganizationProjectAndName>[] = [];
    const promises = streamPaths.map((path, idx) => {
      const parts = path.substring(1, path.length-1).split("/");
      return client
        .query<StreamByOrganizationProjectAndName, StreamByOrganizationProjectAndNameVariables>({
          query: QUERY_STREAM,
          variables: {
            organizationName: toBackendName(parts[0]),
            projectName: toBackendName(parts[1]),
            streamName: toBackendName(parts[2]),
          },
        })
        .then((res) => (results[idx] = res));
    });
    Promise.all(promises).then(() => setStreams(results));
  }, [JSON.stringify(streamPaths)]);

  // Render
  const classes = useStyles();
  return (
    <Grid container spacing={2}>
      {/* Left */}
      <Grid item xs={12} md={8} lg={9}>
        <Grid container spacing={2} direction="column">
          {/* Editor */}
          <Grid item xs={12}>
            <CodeEditor
              rows={15}
              language="sql"
              value={queryText}
              onChange={(value: string) => setQueryText(value)}
            />
          </Grid>
          {/* Action bar */}
          <Grid item xs={12}>
            <Grid container spacing={1}>
              {(loading || (job?.status && job.status !== "done")) && (
                <Grid className={classes.statusAction} item>
                  <Chip label={`Job ${job?.status ?? "pending"}`} />
                </Grid>
              )}
              {job?.bytesScanned && (
                <Grid className={classes.statusAction} item>
                  <Chip
                    label={
                      (!job.jobID ? "Query will scan " : job.status !== "done" ? "Scanning " : "Query scanned ") +
                      numbro(job.bytesScanned).format({
                        base: "binary",
                        mantissa: 1,
                        output: "byte",
                        spaceSeparated: true,
                        optionalMantissa: true,
                        trimMantissa: true,
                      })
                    }
                  />
                </Grid>
              )}
              {job?.resultSizeRecords && (
                <Grid className={classes.statusAction} item>
                  <Chip label={`Result contains ${job?.resultSizeRecords} records`} />
                </Grid>
              )}
              <Grid item xs></Grid>
              {loading && (
                <Grid className={classes.statusAction} item>
                  <CircularProgress size={24} />
                </Grid>
              )}
              <Grid item>
                <Button variant="contained" disabled={loading} onClick={() => analyzeQuery(queryText)}>
                  Analyze
                </Button>
              </Grid>
              <Grid item>
                <Button variant="contained" color="primary" disabled={loading} onClick={() => runQuery(queryText)}>
                  Run
                </Button>
              </Grid>
            </Grid>
          </Grid>
          {/* Table */}
          <Grid item xs={12}>
            <RecordsTable
              schema={schema}
              records={records}
              fetchMore={fetchMore}
              loading={loading}
              error={!!error}
              message={
                error
                  ? error.message
                  : records === undefined
                  ? "Query result will appear here"
                  : records.length === 0
                  ? "Result is empty"
                  : ""
              }
            />
          </Grid>
        </Grid>
      </Grid>
      {/* Right */}
      <Grid item xs={12} md={4} lg={3}>
        <Grid container spacing={2} direction="column">
          {streams.map((result, idx) => (
            <Grid key={idx} item>
              <StreamPreview result={result} />
            </Grid>
          ))}
        </Grid>
      </Grid>
    </Grid>
  );
};

export default Main;
