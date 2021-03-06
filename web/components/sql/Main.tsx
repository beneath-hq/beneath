import { useApolloClient, ApolloQueryResult } from "@apollo/client";
import { Button, CircularProgress, Grid, makeStyles, Theme, Chip, Typography } from "@material-ui/core";
import { useWarehouse } from "beneath-react";
import _ from "lodash";
import { useRouter } from "next/router";
import numbro from "numbro";
import React, { useState, useMemo, useEffect } from "react";

import CodeEditor from "components/CodeEditor";
import RecordsView from "components/table/RecordsView";
import { Schema } from "components/table/schema";
import { useToken } from "hooks/useToken";
import { QUERY_TABLE } from "apollo/queries/table";
import {
  TableByOrganizationProjectAndName,
  TableByOrganizationProjectAndNameVariables,
} from "apollo/types/TableByOrganizationProjectAndName";
import { toBackendName } from "lib/names";
import TablePreview from "./TablePreview";
import { NakedLink } from "components/Link";

const useStyles = makeStyles((theme: Theme) => ({
  unauthenticatedContainer: {
    marginTop: theme.spacing(12),
  },
  statusAction: {
    display: "flex",
    alignItems: "center",
  },
}));

const Main = () => {
  // Prepopulate query text if &table=... url param is set
  const router = useRouter();
  const [queryText, setQueryText] = useState(() => {
    const prepopulateTable = router.query.table;
    if (prepopulateTable) {
      return `select count(*) from \`${prepopulateTable}\``;
    }
    return "";
  });

  // Running query
  const token = useToken();
  const { analyzeQuery, runQuery, loading, error, job, records, fetchMore } = useWarehouse({
    secret: token || undefined,
  });

  // Compute Schema object for job.resultAvroSchema
  const schema = useMemo(() => {
    const schema = job?.resultAvroSchema ? new Schema(job?.resultAvroSchema, []) : undefined;
    return schema;
  }, [job?.resultAvroSchema]);

  // Extract table paths from query
  const tablePaths = useMemo(() => {
    const matches = queryText.match(/`[_\-a-z0-9]+\/[_\-a-z0-9]+\/[_\-a-z0-9]+`/g);
    const paths = _.uniq(matches);
    return paths;
  }, [queryText]);

  // Lookup every table path in query
  const client = useApolloClient();
  const [tables, setTables] = useState<ApolloQueryResult<TableByOrganizationProjectAndName>[]>([]);
  useEffect(() => {
    const results: ApolloQueryResult<TableByOrganizationProjectAndName>[] = [];
    const promises = tablePaths.map((path, idx) => {
      const parts = path.substring(1, path.length - 1).split("/");
      return client
        .query<TableByOrganizationProjectAndName, TableByOrganizationProjectAndNameVariables>({
          query: QUERY_TABLE,
          variables: {
            organizationName: toBackendName(parts[0]),
            projectName: toBackendName(parts[1]),
            tableName: toBackendName(parts[2]),
          },
        })
        .then((res) => (results[idx] = res));
    });
    Promise.all(promises).then(() => setTables(results));
  }, [JSON.stringify(tablePaths)]);

  // get classes
  const classes = useStyles();

  // show message if not logged in
  if (token === null) {
    return (
      <Grid container direction="column" alignItems="center" spacing={2} className={classes.unauthenticatedContainer}>
        <Grid item>
          <Typography variant="h3" gutterBottom>
            Log in or create a free user to run SQL queries
          </Typography>
        </Grid>
        <Grid item>
          <Button component={NakedLink} variant="contained" href="/" color="primary">
            Sign up / Log in
          </Button>
        </Grid>
      </Grid>
    );
  }

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
              onChange={(value: string | undefined) => setQueryText(value || "")}
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
          {/* Records */}
          <Grid item xs={12}>
            <RecordsView
              paper
              schema={schema}
              records={records}
              fetchMore={fetchMore}
              loading={loading}
              error={error?.message}
              callToAction={{
                message:
                  records === undefined
                    ? "Query result will appear here"
                    : records.length === 0
                    ? "Result is empty"
                    : undefined,
              }}
            />
          </Grid>
        </Grid>
      </Grid>
      {/* Right */}
      <Grid item xs={12} md={4} lg={3}>
        <Grid container spacing={2} direction="column">
          {tables.map((result, idx) => (
            <Grid key={idx} item>
              <TablePreview result={result} />
            </Grid>
          ))}
        </Grid>
      </Grid>
    </Grid>
  );
};

export default Main;
