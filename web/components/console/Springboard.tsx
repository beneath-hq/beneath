import { Grid, Typography } from "@material-ui/core";
import React, { FC } from "react";

import useMe from "../../hooks/useMe";
import { toURLName } from "../../lib/names";
import ExploreProjectsTiles from "./ExploreProjectsTiles";
import MyProjectsTiles from "./MyProjectsTiles";
import ActionTile from "./tiles/ActionTile";
import HeroTile from "./tiles/HeroTile";
import UsageTile from "./tiles/UsageTile";
import VSpace from "components/VSpace";

const Springboard: FC = () => {
  const me = useMe();
  if (!me || !me.personalUserID) {
    return <p>Need to log in to view your dashboard -- this shouldn't ever get hit</p>;
  }

  const sections = [
    {
      label: "My profile",
      content: (
        <>
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
        </>
      ),
    },
    {
      label: "Create",
      content: (
        <>
          <ActionTile title="Get Started" href="https://about.beneath.dev/docs/quick-starts/" />
          <ActionTile
            title="Create Secret"
            href={`/organization?organization_name=${me.name}&tab=secrets`}
            as={`/${me.name}/-/secrets`}
          />
          <ActionTile title="Create Project" href={`/-/create/project`} />
          <ActionTile title="Create Stream" href={`/-/create/stream`} />
        </>
      ),
    },
    {
      label: "My projects",
      content: <MyProjectsTiles me={me} />,
    },
    {
      label: "Featured projects",
      content: <ExploreProjectsTiles />,
    },
  ];
  return (
    <Grid container direction="column">
      {sections.map((section) => {
        return (
          <>
            <Grid item>
              <Grid container direction="column" spacing={3}>
                <Grid item>
                  <Typography component="h2" variant="h2">
                    {section.label}
                  </Typography>
                </Grid>
                <Grid item>
                  <Grid container spacing={3}>
                    {section.content}
                  </Grid>
                </Grid>
              </Grid>
            </Grid>
            <VSpace units={10} />
          </>
        );
      })}
    </Grid>
  );
};

export default Springboard;
