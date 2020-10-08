import { FC } from "react";
import { Grid, makeStyles, Typography } from "@material-ui/core";

import Avatar from "../../Avatar";
import { Tile, TileProps } from "./Tile";

const useStyles = makeStyles((theme) => ({
  avatar: {
    paddingRight: theme.spacing(2),
  },
  container: {
    padding: theme.spacing(3),
  },
  orgName: {
    marginRight: theme.spacing(0.75),
  },
  path: {
    [theme.breakpoints.down("md")]: {
      alignItems: "center",
    },
  },
}));

export interface ProjectHeroTileProps extends TileProps {
  name: string;
  organizationName: string;
  displayName?: string | null;
  description?: string | null;
  avatarURL?: string | null;
}

const ProjectHeroTile: FC<ProjectHeroTileProps> = ({
  name,
  organizationName,
  displayName,
  description,
  avatarURL,
  shape,
  ...tileProps
}) => {
  const classes = useStyles();
  return (
    <Tile shape={shape} {...tileProps}>
      <Grid container spacing={2} className={classes.container} alignItems="center">
        <Grid item className={classes.avatar}>
          <Avatar size="list" label={displayName || name} src={avatarURL} />
        </Grid>
        <Grid item xs={9}>
          <Grid container className={classes.path}>
            <Grid item>
              <Typography color="textSecondary" className={classes.orgName}>
                {organizationName} /
              </Typography>
            </Grid>
            <Grid item>
              <Typography variant="h3">{name}</Typography>
            </Grid>
          </Grid>
        </Grid>
        {description && (
          <Grid item xs={12}>
            <Typography variant="body1">
              {description}
            </Typography>
          </Grid>
        )}
      </Grid>
    </Tile>
  );
};

export default ProjectHeroTile;
