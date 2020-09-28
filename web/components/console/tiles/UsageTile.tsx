import React, { FC } from "react";
import { Grid, makeStyles, Theme, Typography } from "@material-ui/core";

import { ActualIndicator } from "../../metrics/user/UsageIndicator";
import { Tile, TileProps } from "./Tile";

const useStyles = makeStyles((theme: Theme) => ({
  container: {
    padding: theme.spacing(2),
  },
}));

export interface UsageTileProps extends TileProps {
  title: string;
  usage: number;
  quota: number;
}

export const UsageTile: FC<UsageTileProps> = ({ title, usage, quota, ...tileProps }) => {
  const classes = useStyles();
  return (
    <Tile {...tileProps}>
      <Grid className={classes.container} container justify="center" alignContent="center" alignItems="center">
        <Grid item xs>
          <Typography variant="h3" gutterBottom>
            {title}
          </Typography>
          <ActualIndicator standalone={false} kind={"read"} usage={usage} quota={quota} />
        </Grid>
      </Grid>
    </Tile>
  );
};

export default UsageTile;
