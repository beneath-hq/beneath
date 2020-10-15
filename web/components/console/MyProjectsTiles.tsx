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
import { BillingInfo, BillingInfoVariables } from "apollo/types/BillingInfo";
import { QUERY_BILLING_INFO } from "apollo/queries/billinginfo";

const MyProjectsTiles: FC = () => {
  const me = useMe();
  const theme = useTheme();
  const isMd = useMediaQuery(theme.breakpoints.up("md"));

  if (!me || !me.personalUserID) {
    return <></>;
  }

  const { loading, error, data } = useQuery<ProjectsForUser, ProjectsForUserVariables>(QUERY_PROJECTS_FOR_USER, {
    fetchPolicy: "cache-and-network",
    variables: {
      userID: me.personalUserID,
    },
  });

  const { loading: loading2, error: error2, data: data2 } = useQuery<BillingInfo, BillingInfoVariables>(QUERY_BILLING_INFO, {
    variables: {
      organizationID: me.organizationID,
    },
  });

  return (
    <>
      {(loading || loading2 ) && <LoadingTile />}
      {error && <ErrorTile error={error?.message || error2?.message || "Couldn't load your projects"} />}
      {data &&
        data.projectsForUser &&
        data.projectsForUser.map(({ projectID, name, description, photoURL, organization }, i) => (
          <React.Fragment key={i}>
            {/* if its the 3rd or 5th item, on medium+ screens, and you're on the free plan, then add a spacer to avoid overlap with the UpgradeTile */}
             {(i === 2 || i === 4) && isMd && data2 && data2.billingInfo.billingPlan.default && <Grid item xs={4} />}
              <ProjectHeroTile
                key={`explore:${projectID}`}
                href={`/project?organization_name=${toURLName(organization.name)}&project_name=${toURLName(name)}`}
                as={`/${toURLName(organization.name)}/${toURLName(name)}`}
                organizationName={organization.name}
                name={toURLName(name)}
                description={description}
                avatarURL={photoURL}
              />
            </React.Fragment>
          )
        )}
      {data && !data.projectsForUser && <PlaceholderTile title="Your first project will show up here" />}
    </>
  );
};

export default MyProjectsTiles;
