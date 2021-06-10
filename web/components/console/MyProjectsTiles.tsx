import { useQuery } from "@apollo/client";
import React, { FC } from "react";
import { Grid, useMediaQuery, useTheme } from "@material-ui/core";

import { QUERY_PROJECTS_FOR_USER } from "../../apollo/queries/project";
import { ProjectsForUser, ProjectsForUserVariables } from "../../apollo/types/ProjectsForUser";
import { toURLName } from "../../lib/names";
import ErrorTile from "./tiles/ErrorTile";
import LoadingTile from "./tiles/LoadingTile";
import ProjectHeroTile from "./tiles/ProjectHeroTile";
import PlaceholderTile from "./tiles/PlaceholderTile";
import useMe from "hooks/useMe";

const MyProjectsTiles: FC = () => {
  const me = useMe();
  const theme = useTheme();
  const isMd = useMediaQuery(theme.breakpoints.up("md"), { defaultMatches: true });

  if (!me || !me.personalUserID) {
    return <></>;
  }

  const { loading, error, data } = useQuery<ProjectsForUser, ProjectsForUserVariables>(QUERY_PROJECTS_FOR_USER, {
    fetchPolicy: "cache-and-network",
    variables: {
      userID: me.personalUserID,
    },
  });

  return (
    <>
      {loading && <LoadingTile />}
      {error && <ErrorTile error={error?.message || "Couldn't load your projects"} />}
      {data &&
        data.projectsForUser &&
        data.projectsForUser.map(({ projectID, name, description, photoURL, public: isPublic, organization }, i) => (
          <React.Fragment key={i}>
            {/* if its the 3rd or 5th item, on medium+ screens, then add a spacer to avoid overlap with the Metrics panel */}
            {(i === 2 || i === 4) && isMd && <Grid item xs={4} />}
            <ProjectHeroTile
              key={`explore:${projectID}`}
              href={`/project?organization_name=${toURLName(organization.name)}&project_name=${toURLName(name)}`}
              as={`/${toURLName(organization.name)}/${toURLName(name)}`}
              organizationName={organization.name}
              name={toURLName(name)}
              description={description}
              avatarURL={photoURL}
              isPublic={isPublic}
            />
          </React.Fragment>
        ))}
      {data && !data.projectsForUser && <PlaceholderTile title="Your first project will show up here" />}
    </>
  );
};

export default MyProjectsTiles;
