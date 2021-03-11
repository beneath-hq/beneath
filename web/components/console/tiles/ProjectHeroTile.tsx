import { FC } from "react";
import { Chip, Grid, makeStyles, Tooltip, Typography } from "@material-ui/core";

import Avatar from "../../Avatar";
import { Tile, TileProps } from "./Tile";
import { toURLName } from "lib/names";
import clsx from "clsx";

const useStyles = makeStyles((theme) => ({
  container: {
    padding: theme.spacing(2),
  },
  orgName: {
    fontSize: theme.typography.caption.fontSize,
  },
  pointer: {
    cursor: "pointer",
  },
  publicChip: {
    background: theme.palette.primary.dark,
  },
}));

export interface ProjectHeroTileProps extends TileProps {
  name: string;
  organizationName: string;
  displayName?: string | null;
  description?: string | null;
  avatarURL?: string | null;
  isPublic: boolean;
}

const ProjectHeroTile: FC<ProjectHeroTileProps> = ({
  name,
  organizationName,
  displayName,
  description,
  avatarURL,
  isPublic,
  shape,
  ...tileProps
}) => {
  const classes = useStyles();
  return (
    <Tile shape={shape} {...tileProps}>
      <Grid container spacing={2} className={classes.container}>
        <Grid item>
          <Avatar size="list" label={displayName || name} src={avatarURL} />
        </Grid>
        <Grid item xs>
          <Grid container direction="column">
            <Grid item>
              <Grid container alignItems="flex-start">
                <Grid item>
                  <Typography color="textSecondary" className={classes.orgName}>
                    {toURLName(organizationName)} /
                  </Typography>
                </Grid>
                <Grid item xs />
                <Grid item>
                  <Tooltip
                    title={
                      isPublic ? "Anyone can access this project" : "People need permission to access this project"
                    }
                  >
                    <Chip
                      label={isPublic ? "Public" : "Private"}
                      size="small"
                      className={clsx(classes.pointer, isPublic && classes.publicChip)}
                    />
                  </Tooltip>
                </Grid>
              </Grid>
            </Grid>
            <Grid item>
              <Typography variant="h3">{toURLName(name)}</Typography>
            </Grid>
          </Grid>
        </Grid>
        {description && (
          <Grid item xs={12}>
            <Typography variant="body1">{description}</Typography>
          </Grid>
        )}
      </Grid>
    </Tile>
  );
};

export default ProjectHeroTile;
