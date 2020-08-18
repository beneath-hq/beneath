import CircularProgress from "@material-ui/core/CircularProgress";
import Grid, { GridJustification } from "@material-ui/core/Grid";
import { FC } from "react";

export interface LoadingProps {
  justify?: GridJustification;
  size?: number | string;
}

export const Loading: FC<LoadingProps> = (props) => (
  <Grid container justify={props.justify}>
    <Grid item>
      <CircularProgress size={props.size} />
    </Grid>
  </Grid>
);

export default Loading;
