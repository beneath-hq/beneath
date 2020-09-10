import { useApolloClient, ApolloQueryResult } from "@apollo/client";
import { Button, CircularProgress, Grid, makeStyles, Theme, Typography, Paper } from "@material-ui/core";
import { useWarehouse } from "beneath-react";
import _ from "lodash";
import { useRouter } from "next/router";
import React, { useState, useMemo, useEffect, FC } from "react";

import TextField from "../forms/TextField";
import { Link } from "components/Link";
import RecordsTable from "components/stream/RecordsTable";
import { Schema } from "components/stream/schema";
import { useToken } from "hooks/useToken";
import { QUERY_STREAM } from "apollo/queries/stream";
import { StreamByOrganizationProjectAndName, StreamByOrganizationProjectAndNameVariables } from "apollo/types/StreamByOrganizationProjectAndName";
import { toBackendName, toURLName } from "lib/names";
import SchemaView from "./SchemaView";

const useStyles = makeStyles((_: Theme) => ({
  statusAction: {
    display: "flex",
    alignItems: "center",
  },
  streamPreviewPaper: {
    padding: "12px",
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
      <Grid item xs={12} md={8} container spacing={2}>
        {/* Editor */}
        <Grid item xs={12}>
          <TextField
            multiline
            margin="none"
            monospace
            rows={15}
            rowsMax={200}
            value={queryText}
            onChange={(e) => setQueryText(e.target.value)}
          />
        </Grid>
        {/* Action bar */}
        <Grid item xs={12} container spacing={1}>
          {loading && (
            <Grid className={classes.statusAction} item>
              <CircularProgress size={28} />
            </Grid>
          )}
          {(loading || (job?.status && job.status !== "done")) && (
            <Grid className={classes.statusAction} item>
              <Typography>Job status: {job?.status ?? "pending"}</Typography>
            </Grid>
          )}
          {job?.bytesScanned && (
            <Grid className={classes.statusAction} item>
              <Typography>Query scans {job?.bytesScanned} bytes</Typography>
            </Grid>
          )}
          <Grid item xs></Grid>
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
                : !job
                ? "Query result will appear here"
                : !job.jobID
                ? "Query result will appear here"
                : ""
            }
          />
        </Grid>
      </Grid>
      {/* Right */}
      <Grid item xs={12} md={4} container spacing={2}>
        {streams.map((result, idx) => (
          <Grid key={idx} item xs={12}>
            <StreamPreview result={result} />
          </Grid>
        ))}
      </Grid>
    </Grid>
  );
};

export default Main;

interface StreamPreviewProps {
  result: ApolloQueryResult<StreamByOrganizationProjectAndName>;
}

const StreamPreview: FC<StreamPreviewProps> = ({ result: { data, error, errors } }) => {
  let errorMsg = null;
  if (error || errors || !data) {
    errorMsg = error ? error.message : (errors && errors.length > 0) ? errors[0].message : "Unknown error";
  }

  const stream = data?.streamByOrganizationProjectAndName;
  const orgName = toURLName(stream?.project.organization.name || "");
  const projName = toURLName(stream?.project.name || "");

  const classes = useStyles();
  return (
    <Paper className={classes.streamPreviewPaper}>
      {errorMsg && <Typography>{errorMsg}</Typography>}
      {stream && (
        <>
          <Link
            color="textPrimary"
            underline="none"
            variant="h3"
            gutterBottom
            href={`/stream?organization_name=${orgName}&project_name=${projName}&stream_name=${stream.name}`}
            as={`/${orgName}/${projName}/${stream.name}`}
          >
            {orgName}/{projName}/{stream.name}
          </Link>
          <Typography color="textSecondary" gutterBottom>
            {stream.description}
          </Typography>
          <SchemaView stream={stream} />
        </>
      )}
    </Paper>
  );
};
