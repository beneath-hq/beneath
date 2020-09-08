import { Button, CircularProgress, Grid, makeStyles, Theme, Typography } from "@material-ui/core";
import { useWarehouse } from "beneath-react";
import { useRouter } from "next/router";
import React, { useState, useMemo } from "react";

import TextField from "../forms/TextField";
import RecordsTable from "components/stream/RecordsTable";
import { Schema } from "components/stream/schema";
import { useToken } from "hooks/useToken";

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

  const schema = useMemo(() => {
    const schema = job?.resultAvroSchema ? new Schema(job?.resultAvroSchema, []) : undefined;
    return schema;
  }, [job?.resultAvroSchema]);

  const classes = useStyles();
  return (
    <Grid container spacing={4}>
      {/* Left */}
      <Grid item xs={12} md={8} container spacing={1}>
        {/* Editor */}
        <Grid item xs={12}>
          <TextField
            multiline
            monospace
            rows={15}
            rowsMax={200}
            value={queryText}
            onChange={(e) => setQueryText(e.target.value)}
          />
        </Grid>
        {/* Action bar */}
        <Grid item xs={12} container spacing={1}>
          {loading && <Grid className={classes.statusAction} item><CircularProgress size={28} /></Grid>}
          {(loading || job?.status && job.status !== "done") && <Grid className={classes.statusAction} item><Typography>Job status: {job?.status ?? "pending"}</Typography></Grid>}
          {job?.bytesScanned && <Grid className={classes.statusAction} item><Typography>Query scans {job?.bytesScanned} bytes</Typography></Grid>}
          <Grid item xs></Grid>
          <Grid item>
            <Button disabled={loading} onClick={() => analyzeQuery(queryText)}>
              Analyze
            </Button>
          </Grid>
          <Grid item>
            <Button color="primary" disabled={loading} onClick={() => runQuery(queryText)}>
              Run
            </Button>
          </Grid>
        </Grid>
        {/* Table */}
        <Grid item xs={12}>
          {schema && records && <RecordsTable schema={schema} records={records} />}
          {error && <p>Error: {JSON.stringify(error.message)}</p>}
          {fetchMore && <Button onClick={() => fetchMore()}>Fetch more</Button>}
        </Grid>
      </Grid>
      {/* Right */}
      <Grid item xs={12} md={4} container spacing={1}>
        <Grid item xs={12}>
          <TextField multiline rows={10} rowsMax={200} />
        </Grid>
      </Grid>
    </Grid>
  );
};

export default Main;
