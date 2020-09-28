import { useQuery } from "@apollo/client";
import { FC } from "react";

import { QUERY_PROJECTS_FOR_USER } from "../../apollo/queries/project";
import { Me_me } from "../../apollo/types/Me";
import { ProjectsForUser, ProjectsForUserVariables } from "../../apollo/types/ProjectsForUser";
import { toURLName } from "../../lib/names";
import ErrorTile from "./tiles/ErrorTile";
import LoadingTile from "./tiles/LoadingTile";
import ProjectHeroTile from "./tiles/ProjectHeroTile";

export interface MyProjectsTilesProps {
  me: Me_me;
}

const MyProjectsTiles: FC<MyProjectsTilesProps> = ({ me }) => {
  if (!me.personalUserID) {
    return <></>;
  }

  const { loading, error, data } = useQuery<ProjectsForUser, ProjectsForUserVariables>(QUERY_PROJECTS_FOR_USER, {
    variables: {
      userID: me.personalUserID,
    },
  });

  return (
    <>
      {loading && <LoadingTile />}
      {error && <ErrorTile error={error?.message || "Couldn't load your projects"} />}
      {data && data.projectsForUser &&
        data.projectsForUser.map(({ projectID, name, description, photoURL, organization }) => (
          <ProjectHeroTile
            key={`explore:${projectID}`}
            href={`/project?organization_name=${toURLName(organization.name)}&project_name=${toURLName(name)}`}
            as={`/${toURLName(organization.name)}/${toURLName(name)}`}
            organizationName={organization.name}
            name={toURLName(name)}
            description={description}
            avatarURL={photoURL}
          />
        ))}
    </>
  );
};

export default MyProjectsTiles;
