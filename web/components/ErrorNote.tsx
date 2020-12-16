import { Grid, Typography } from "@material-ui/core";
import { FC } from "react";

const ErrorNote: FC<{ error?: Error }> = ({ error }) => {
  return (
    <Grid item xs={12}>
      <Typography color="error">{error ? error.message : "Unexpected error"}</Typography>
    </Grid>
  );
};

export default ErrorNote;
