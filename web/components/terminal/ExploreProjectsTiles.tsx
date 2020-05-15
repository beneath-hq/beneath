import { useQuery } from "@apollo/react-hooks";
import { FC } from "react";

import {
  makeStyles,
  Theme,
} from "@material-ui/core";

import { EXPLORE_PROJECTS } from "../../apollo/queries/project";
import { ExploreProjects } from "../../apollo/types/ExploreProjects";
import { toURLName } from "../../lib/names";

import ErrorTile from "./tiles/ErrorTile";
import HeroTile from "./tiles/HeroTile";
import LoadingTile from "./tiles/LoadingTile";

const useStyles = makeStyles((theme: Theme) => ({
}));

const ExploreProjectsTiles: FC = () => {
  const { loading, error, data } = useQuery<ExploreProjects>(EXPLORE_PROJECTS);
  const classes = useStyles();
  return (
    <>
      {loading && <LoadingTile />}
      {error && <ErrorTile error={error?.message || "Couldn't load projects to explore"} />}
      {data && data.exploreProjects &&
        data.exploreProjects.map(({ projectID, name, displayName, description, photoURL, organization }) => (
          <HeroTile
            key={`explore:${projectID}`}
            href={`/project?organization_name=${toURLName(organization.name)}&project_name=${toURLName(name)}`}
            as={`/${toURLName(organization.name)}/${toURLName(name)}`}
            path={`/${toURLName(organization.name)}/${toURLName(name)}`}
            name={toURLName(name)}
            displayName={displayName}
            description={description}
            avatarURL={photoURL}
          />
        ))}
    </>
  );
};

export default ExploreProjectsTiles;
