import { Grid, makeStyles, Theme, Typography, useMediaQuery, useTheme } from "@material-ui/core";
import React, { FC } from "react";

import useMe from "../../hooks/useMe";
import { toURLName } from "../../lib/names";
import ExploreProjectsTiles from "./ExploreProjectsTiles";
import MyProjectsTiles from "./MyProjectsTiles";
import MyUsageTile from "./tiles/MyUsageTile";
import ProfileHeroTile from "./tiles/ProfileHeroTile";
import UpgradeTile from "./tiles/UpgradeTile";
import ActionsTile from "./tiles/ActionsTile";
import ContentContainer from "components/ContentContainer";

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
  usageTile: {
    [theme.breakpoints.up("md")]: {
      position: "absolute",
      right: 0,
      top: 0,
      width: 300,
    },
  },
  upgradeTile: {
    [theme.breakpoints.up("md")]: {
      position: "absolute",
      right: 0,
      top: 340,
      width: 300,
    },
  },
}));

const Springboard: FC = () => {
  const me = useMe();
  const classes = useStyles();
  const theme = useTheme();
  const isMd = useMediaQuery(theme.breakpoints.up("md"), { defaultMatches: true });

  if (!me || !me.personalUserID) {
    return <p>Need to log in to view your dashboard -- this shouldn't ever get hit</p>;
  }

  return (
    <ContentContainer maxWidth="lg">
      <Grid container spacing={3} className={classes.positionAncestor}>
        <ProfileHeroTile
          shape="wide"
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
            <MyUsageTile className={classes.usageTile} />
            <UpgradeTile className={classes.upgradeTile} />
          </>
        )}
        {isMd && <Grid item md={4} lg={6} />}
        <ActionsTile shape="wide" nopaper />
        <Grid item xs={12}>
          <Typography variant="h3" className={classes.sectionTitle}>
            My projects
          </Typography>
        </Grid>
        <MyProjectsTiles />
        <Grid item xs={12}>
          <Typography variant="h3" className={classes.sectionTitle}>
            Featured projects
          </Typography>
        </Grid>
        <ExploreProjectsTiles />
      </Grid>
    </ContentContainer>
  );
};

export default Springboard;
