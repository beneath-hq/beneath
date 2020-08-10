import { Grid, makeStyles, Theme } from "@material-ui/core";
import React, { FC } from "react";

import useMe from "../../hooks/useMe";
import { toURLName } from "../../lib/names";
import ExploreProjectsTiles from "./ExploreProjectsTiles";
import MyProjectsTiles from "./MyProjectsTiles";
import ActionTile from "./tiles/ActionTile";
import HeroTile from "./tiles/HeroTile";
import TitleTile from "./tiles/TitleTile";
import UsageTile from "./tiles/UsageTile";

const useStyles = makeStyles((theme: Theme) => ({
}));

const Springboard: FC = () => {
  const classes = useStyles();

  const me = useMe();
  if (!me || !me.personalUserID) {
    return <p>Need to log in to view your dashboard -- this shouldn't ever get hit</p>;
  }

  return (
    <Grid container spacing={3} alignItems="stretch" alignContent="stretch">
      <HeroTile
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
          {me.readQuota && (
            <UsageTile
              href={`/organization?organization_name=${me.name}&tab=monitoring`}
              as={`/${me.name}/-/monitoring`}
              title="Read quota usage"
              usage={me.readUsage}
              quota={me.readQuota}
            />
          )}
          {me.writeQuota && (
            <UsageTile
              href={`/organization?organization_name=${me.name}&tab=monitoring`}
              as={`/${me.name}/-/monitoring`}
              title="Write quota usage"
              usage={me.writeUsage}
              quota={me.writeQuota}
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
      <ActionTile title="Get Started" href="https://about.beneath.dev/docs/quick-starts/" />
      <ActionTile
        title="Create Secret"
        href={`/organization?organization_name=${me.name}&tab=secrets`}
        as={`/${me.name}/-/secrets`}
      />
      <MyProjectsTiles me={me} />
      <TitleTile title="Featured projects: " />
      <ExploreProjectsTiles />
    </Grid>
  );
};

export default Springboard;
