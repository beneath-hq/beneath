import React, { FC } from "react";
import { Query } from "react-apollo";

import { makeStyles, Theme } from "@material-ui/core";
import Button from "@material-ui/core/Button";
import Grid from "@material-ui/core/Grid";

import { QUERY_RECORDS } from "../../apollo/queries/local/records";
import { QueryStream } from "../../apollo/types/QueryStream";
import { Records, RecordsVariables } from "../../apollo/types/Records";
import BNTextField from "../BNTextField";
import Loading from "../Loading";
import RecordsTable from "../RecordsTable";
import VSpace from "../VSpace";
import { Schema } from "./schema";

const useStyles = makeStyles((theme: Theme) => ({
  submitButton: {
    marginTop: theme.spacing(3),
  },
}));

const ExploreStream: FC<QueryStream> = ({ stream }) => {
  const schema = new Schema(stream, false);

  const [values, setValues] = React.useState({
    where: "",
    vars: {
      projectName: stream.project.name,
      streamName: stream.name,
      limit: 100,
      where: "",
    } as RecordsVariables,
  });

  const handleChange = (name: string) => (event: any) => {
    setValues({ ...values, [name]: event.target.value });
  };

  const classes = useStyles();
  return (
    <Query<Records, RecordsVariables> query={QUERY_RECORDS} variables={values.vars} fetchPolicy="cache-and-network">
      {({ loading, error, data }) => {
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
                  helperText="Query the stream on indexed fields"
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
                  Load
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

        return (
          <>
            {formElem}
            <VSpace units={2} />
            {tableElem}
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
