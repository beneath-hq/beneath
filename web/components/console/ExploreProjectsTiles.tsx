import { useQuery } from "@apollo/client";
import { FC } from "react";

import { EXPLORE_PROJECTS } from "../../apollo/queries/project";
import { ExploreProjects } from "../../apollo/types/ExploreProjects";
import { toURLName } from "../../lib/names";
import ErrorTile from "./tiles/ErrorTile";
import LoadingTile from "./tiles/LoadingTile";
import ProjectHeroTile from "./tiles/ProjectHeroTile";

const ExploreProjectsTiles: FC = () => {
  const { loading, error, data } = useQuery<ExploreProjects>(EXPLORE_PROJECTS);
  return (
    <>
      {loading && <LoadingTile />}
      {error && <ErrorTile error={error?.message || "Couldn't load projects to explore"} />}
      {data &&
        data.exploreProjects &&
        data.exploreProjects.map(({ projectID, name, description, photoURL, public: isPublic, organization }) => (
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
        ))}
    </>
  );
};

export default ExploreProjectsTiles;
