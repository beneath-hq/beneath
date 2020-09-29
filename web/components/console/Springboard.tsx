import { Grid, makeStyles, Theme, Typography, Link } from "@material-ui/core";
import React, { FC } from "react";
import { ArrowRightAlt, Folder, LinearScale, VpnKey } from "@material-ui/icons";

import useMe from "../../hooks/useMe";
import { toURLName } from "../../lib/names";
import ExploreProjectsTiles from "./ExploreProjectsTiles";
import MyProjectsTiles from "./MyProjectsTiles";
import UsageTile from "./tiles/UsageTile";
import ActionTile from "./tiles/ActionTile";
import ProfileHeroTile from "./tiles/ProfileHeroTile";

// Hack: in order to position the usage section in the top right-hand corner on medium+ screens,
// we apply custom css styling to the Grid items. We pass through the styles to the Tile component, which is itself a Grid item.
// Note that there's buggy behavior when we nest an extra Grid layer on top of the current structure, so we avoid that.

const useStyles = makeStyles((theme: Theme) => ({
  sectionTitle: {
    marginTop: theme.spacing(4),
  },
  positionAncestor: {
    [theme.breakpoints.up("md")]: {
      position: "relative",
    },
  },
  usageTitle: {
    [theme.breakpoints.up("md")]: {
      position: "absolute",
      right: 0,
      top: 0,
      width: 300
    },
    [theme.breakpoints.down("sm")]: {
      marginTop: theme.spacing(4),
    },
  },
  readUsage: {
    [theme.breakpoints.up("md")]: {
      position: "absolute",
      right: 0,
      top: 46,
      width: 300
    },
  },
  writeUsage: {
    [theme.breakpoints.up("md")]: {
      position: "absolute",
      right: 0,
      top: 170,
      width: 300
    },
  },
  usageUpgrade: {
    [theme.breakpoints.up("md")]: {
      position: "absolute",
      right: 0,
      top: 294,
      width: 300
    },
  },
}));

const Springboard: FC = () => {
  const me = useMe();
  const classes = useStyles();

  if (!me || !me.personalUserID) {
    return <p>Need to log in to view your dashboard -- this shouldn't ever get hit</p>;
  }

  return (
    <Grid container spacing={3} className={classes.positionAncestor}>
      <Grid item xs={12}>
        <Typography variant="h3">My profile</Typography>
      </Grid>
      <ProfileHeroTile
        shape="normal"
        href={`/organization?organization_name=${toURLName(me.name)}`}
        as={`/${toURLName(me.name)}`}
        path={`@${toURLName(me.name)}`}
        name={toURLName(me.name)}
        displayName={me.displayName}
        description={me.description}
        avatarURL={me.photoURL}
      />

      {me.organizationID === me.personalUser?.billingOrganizationID && (
        <>
          <Grid item xs={12} className={classes.usageTitle}>
            <Typography variant="h3">
              Usage monitor
            </Typography>
          </Grid>
          {me.readQuota && (
            <UsageTile
              href={`/organization?organization_name=${me.name}&tab=monitoring`}
              as={`/${me.name}/-/monitoring`}
              title="Read quota usage"
              usage={me.readUsage}
              quota={me.readQuota}
              responsive={false}
              styleClass={classes.readUsage}
            />
          )}
          {me.writeQuota && (
            <UsageTile
              href={`/organization?organization_name=${me.name}&tab=monitoring`}
              as={`/${me.name}/-/monitoring`}
              title="Write quota usage"
              usage={me.writeUsage}
              quota={me.writeQuota}
              responsive={false}
              styleClass={classes.writeUsage}
            />
          )}
          {/* {me.scanQuota && (
            <UsageTile
              href={`/organization?organization_name=${me.name}&tab=monitoring`}
              as={`/${me.name}/-/monitoring`}
              title="Scan quota usage"
              usage={me.scanUsage}
              quota={me.scanQuota}
            />
          )} */}
        </>
      )}
      <Grid item xs={12} className={classes.usageUpgrade}>
        <Link href={`/`}>
          <Grid container spacing={1} alignItems="center">
            <Grid item>
              <Typography variant="h3">Upgrade quotas</Typography>
            </Grid>
            <Grid item>
              <ArrowRightAlt />
            </Grid>
          </Grid>
        </Link>
      </Grid>

      <Grid item xs={12}>
        <Typography variant="h3" className={classes.sectionTitle}>
          Create new
        </Typography>
      </Grid>
      <ActionTile title="Project" href={`/-/create/project`} shape="dense">
        <Folder color="primary" />
      </ActionTile>
      <ActionTile title="Stream" href={`/-/create/stream`} shape="dense">
        <LinearScale color="primary" />
      </ActionTile>
      <ActionTile
        title="Secret"
        href={`/organization?organization_name=${me.name}&tab=secrets`}
        as={`/${me.name}/-/secrets`}
        shape="dense"
      >
        <VpnKey color="primary" />
      </ActionTile>

      <Grid item xs={12}>
        <Link href="https://about.beneath.dev/docs/quick-starts/">
          <Grid container spacing={1} alignItems="center">
            <Grid item>
              <Typography variant="h3">Read the quick start</Typography>
            </Grid>
            <Grid item>
              <ArrowRightAlt />
            </Grid>
          </Grid>
        </Link>
      </Grid>

      <Grid item xs={12}>
        <Typography variant="h3" className={classes.sectionTitle}>
          My projects
        </Typography>
      </Grid>
      <MyProjectsTiles me={me} />
      <Grid item xs={12}>
        <Typography variant="h3" className={classes.sectionTitle}>
          Featured projects and tutorials
        </Typography>
      </Grid>
      <ExploreProjectsTiles />
    </Grid>
  );
};

export default Springboard;
