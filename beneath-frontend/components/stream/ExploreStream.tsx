import React, { FC } from "react";
import { Query } from "react-apollo";

import { makeStyles, Theme } from "@material-ui/core";
import Button from "@material-ui/core/Button";
import Grid from "@material-ui/core/Grid";
import Typography from "@material-ui/core/Typography";

import { QUERY_RECORDS } from "../../apollo/queries/local/records";
import { QueryStream } from "../../apollo/types/QueryStream";
import { Records, RecordsVariables } from "../../apollo/types/Records";
import BNTextField from "../BNTextField";
import LinkTypography from "../LinkTypography";
import Loading from "../Loading";
import RecordsTable from "../RecordsTable";
import VSpace from "../VSpace";
import { Schema } from "./schema";

const useStyles = makeStyles((theme: Theme) => ({
  submitButton: {
    marginTop: theme.spacing(1.5),
  },
  fetchMoreButton: {},
  noMoreDataCaption: {
    color: theme.palette.text.disabled,
  },
}));

const ExploreStream: FC<QueryStream> = ({ stream }) => {
  const schema = new Schema(stream, false);

  const [values, setValues] = React.useState({
    where: "",
    noMore: false,
    vars: {
      projectName: stream.project.name,
      streamName: stream.name,
      limit: 20,
      where: "",
    } as RecordsVariables,
  });

  const handleChange = (name: string) => (event: any) => {
    setValues({ ...values, [name]: event.target.value });
  };

  const classes = useStyles();
  return (
    <Query<Records, RecordsVariables> query={QUERY_RECORDS} variables={values.vars} fetchPolicy="cache-and-network">
      {({ loading, error, data, fetchMore }) => {
        const errorMsg = loading ? null : error ? error.message : data ? data.records.error : null;

        const formElem = (
          <form
            onSubmit={(e) => {
              e.preventDefault();
              const where = values.where ? JSON.parse(values.where) : "";
              setValues({ ...values, vars: { ...values.vars, where } });
            }}
          >
            <Grid container spacing={2}>
              <Grid item xs>
                <BNTextField
                  id="where"
                  label="Query"
                  value={values.where}
                  margin="none"
                  onChange={handleChange("where")}
                  helperText={<>
                      You can query the stream on indexed fields, check out the{" "}
                      <LinkTypography href="https://about.beneath.network/docs/using-the-explore-tab/">
                        docs
                      </LinkTypography>{" "}
                      for more info.
                  </>}
                  errorText={errorMsg ? `Error: ${errorMsg}` : undefined}
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
                    loading || !(isJSON(values.where) || values.where.length === 0) || !(values.where.length <= 1024)
                  }
                >
                  Execute
                </Button>
              </Grid>
            </Grid>
          </form>
        );

        let tableElem = null;
        if (loading && !(data && data.records)) {
          tableElem = <Loading justify="center" />;
        } else {
          tableElem = <RecordsTable schema={schema} records={data ? data.records.data : null} />;
        }

        let moreElem = null;
        const records = data && data.records && data.records.data;
        if (tableElem && records && records.length > 0 && records.length % values.vars.limit === 0) {
          moreElem = (
            <Grid container justify="center">
              <Grid item>
                <Button
                  variant="outlined"
                  color="primary"
                  className={classes.fetchMoreButton}
                  disabled={loading}
                  onClick={() => {
                    const after = {} as any;
                    for (const field of schema.keyFields) {
                      after[field] = records[records.length - 1].data[field];
                    }
                    fetchMore({
                      variables: { ...values.vars, after },
                      updateQuery: (prev, { fetchMoreResult }) => {
                        if (!fetchMoreResult) {
                          return prev;
                        }
                        const prevRecords = prev.records.data;
                        const newRecords = fetchMoreResult.records.data;
                        if (!newRecords || newRecords.length === 0) {
                          setValues({ ...values, noMore: true });
                        }
                        if (!prevRecords || !newRecords) {
                          return prev;
                        }
                        return {
                          records: {
                            __typename: "RecordsResponse",
                            data: [...prevRecords, ...newRecords],
                            error: null,
                          },
                        };
                      },
                    });
                  }}
                >
                  Fetch more
                </Button>
              </Grid>
            </Grid>
          );
        }

        return (
          <>
            {formElem}
            <VSpace units={2} />
            {tableElem}
            <VSpace units={4} />
            {!values.noMore && moreElem}
            {(!moreElem || values.noMore) && (
              <Typography className={classes.noMoreDataCaption} variant="body2" align="center">
                There's no more data to load in this stream
              </Typography>
            )}
            <VSpace units={8} />
          </>
        );
      }}
    </Query>
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
