import { useQuery } from "@apollo/react-hooks";
import { FC } from "react";

import {
  makeStyles,
  Theme,
} from "@material-ui/core";

import { QUERY_PROJECTS_FOR_USER } from "../../apollo/queries/project";
import { Me_me } from "../../apollo/types/Me";
import { ProjectsForUser, ProjectsForUserVariables } from "../../apollo/types/ProjectsForUser";
import { toURLName } from "../../lib/names";
import ErrorTile from "./tiles/ErrorTile";
import HeroTile from "./tiles/HeroTile";
import LoadingTile from "./tiles/LoadingTile";

const useStyles = makeStyles((theme: Theme) => ({
}));

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
  const classes = useStyles();
  return (
    <>
      {loading && <LoadingTile />}
      {(error || !data) && <ErrorTile error={error?.message || "Couldn't load your projects"} />}
      {data && data.projectsForUser &&
        data.projectsForUser.map(({ projectID, name, displayName, description, photoURL, organization }) => (
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

export default MyProjectsTiles;
