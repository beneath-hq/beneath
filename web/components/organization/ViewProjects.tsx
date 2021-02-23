import { Grid } from "@material-ui/core";
import React, { FC } from "react";

import { OrganizationByName_organizationByName } from "apollo/types/OrganizationByName";
import ContentContainer, { CallToAction } from "components/ContentContainer";
import { toURLName } from "lib/names";
import ProjectHeroTile from "components/console/tiles/ProjectHeroTile";

export interface ViewProjectsProps {
  organization: OrganizationByName_organizationByName;
}

const ViewProjects: FC<ViewProjectsProps> = ({ organization }) => {
  let cta: CallToAction | undefined;
  if (!organization.projects?.length) {
    cta = {
      message: `We didn't find any projects for @${organization.name}`,
    };
    if (organization.__typename === "PrivateOrganization" && organization.permissions.create) {
      cta.buttons = [{ label: "Create project", href: "/-/create/project" }];
    }
  }

  return (
    <ContentContainer callToAction={cta}>
      <Grid container spacing={3}>
        {organization.projects.map(({ projectID, name, displayName, description, photoURL, public: isPublic }) => (
          <ProjectHeroTile
            key={projectID}
            href={`/project?organization_name=${toURLName(organization.name)}&project_name=${toURLName(name)}`}
            as={`/${toURLName(organization.name)}/${toURLName(name)}`}
            organizationName={organization.name}
            name={name}
            displayName={displayName}
            description={description}
            avatarURL={photoURL}
            isPublic={isPublic}
          />
        ))}
      </Grid>
    </ContentContainer>
  );
};

export default ViewProjects;
