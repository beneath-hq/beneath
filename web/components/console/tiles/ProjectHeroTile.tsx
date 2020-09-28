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
    marginRight: theme.spacing(1),
  },
  description: {
    marginTop: theme.spacing(3),
  },
}));

export interface ProjectHeroTileProps extends TileProps {
  name: string;
  organizationName: string;
  displayName?: string | null;
  path?: string | null;
  description?: string | null;
  avatarURL?: string | null;
}

const ProjectHeroTile: FC<ProjectHeroTileProps> = ({
  name,
  organizationName,
  displayName,
  path,
  description,
  avatarURL,
  shape,
  ...tileProps
}) => {
  const classes = useStyles();
  return (
    <Tile shape={shape} {...tileProps}>
      <Grid container spacing={0} className={classes.container} alignItems="center" xs={12}>
        <Grid item className={classes.avatar}>
          <Avatar size="list" label={displayName || name} src={avatarURL} />
        </Grid>
        <Grid item xs={9}>
          <Grid container>
            <Grid item>
              <Typography color="textSecondary" className={classes.orgName}>
                {organizationName} /
              </Typography>
            </Grid>
            <Grid item>
              <Typography variant="h3">
                {name}
              </Typography>
            </Grid>
          </Grid>
        </Grid>
        {description && (
          <Grid item xs={12}>
            <Typography className={classes.description} variant="body1">
              {description}
            </Typography>
          </Grid>
        )}
      </Grid>
    </Tile>
  );
};

export default ProjectHeroTile;
