import { Grid, Typography } from "@material-ui/core";
import { FC } from "react";

const ErrorNote: FC<{ error: Error }> = ({ error }) => {
  return (
    <Grid item xs={12}>
      <Typography color="error">{error.message}</Typography>
    </Grid>
  );
};

export default ErrorNote;
