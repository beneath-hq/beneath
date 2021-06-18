import { useLazyQuery } from "@apollo/client";
import { Grid, makeStyles, Typography } from "@material-ui/core";
import { FC, useEffect, useState } from "react";

import { COMPILE_SCHEMA } from "apollo/queries/table";
import { CompileSchema, CompileSchemaVariables } from "apollo/types/CompileSchema";
import { TableSchemaKind } from "apollo/types/globalTypes";

const useStyles = makeStyles((theme) => ({
  success: {
    color: theme.palette.success.main,
  },
}));

const MILLISECONDS_TO_WAIT = 2000;

interface Props {
  schemaKind: TableSchemaKind;
  schema: string;
}

const SchemaEditorFooter: FC<Props> = ({ schemaKind, schema }) => {
  const classes = useStyles();
  const [isWaiting, setIsWaiting] = useState(true);
  const [compileSchema, { loading, data, error }] = useLazyQuery<CompileSchema, CompileSchemaVariables>(COMPILE_SCHEMA);

  // when the schema changes, wait X seconds before calling compileSchema()
  // the timer is reset (via clearTimeout) every time the schema changes
  useEffect(() => {
    setIsWaiting(true);
    const timeout = setTimeout(() => {
      if (schema && schema !== "") {
        compileSchema({ variables: { input: { schemaKind, schema } } });
        setIsWaiting(false);
      }
    }, MILLISECONDS_TO_WAIT);
    return () => clearTimeout(timeout);
  }, [schema]);

  const prettyKey = (canonicalIndexes: string) => {
    const objs = JSON.parse(canonicalIndexes);
    const fields = objs[0].fields; // TEMPORARY: assuming only 1 index
    if (fields.length === 1) {
      return fields;
    }
    return "(" + fields.join(", ") + ")";
  };

  return (
    <>
      {!isWaiting && (
        <Grid container justify="space-between">
          <Grid item>
            {data && <Typography>Key: {prettyKey(data.compileSchema.canonicalIndexes)}</Typography>}
            {error && <Typography color="error">{error.message}</Typography>}
          </Grid>
          <Grid item>
            {data && <Typography className={classes.success}>Valid schema</Typography>}
            {loading && <Typography>Compiling...</Typography>}
          </Grid>
        </Grid>
      )}
    </>
  );
};

export default SchemaEditorFooter;
