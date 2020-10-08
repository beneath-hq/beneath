import { List, ListItem, ListItemAvatar, ListItemText } from "@material-ui/core";
import React, { FC } from "react";

import { ProjectByOrganizationAndName_projectByOrganizationAndName } from "apollo/types/ProjectByOrganizationAndName";
import Avatar from "components/Avatar";
import ContentContainer, { CallToAction } from "components/ContentContainer";
import { NakedLink } from "components/Link";
import { toURLName } from "lib/names";

interface ViewStreamsProps {
  project: ProjectByOrganizationAndName_projectByOrganizationAndName;
}

const ViewStreams: FC<ViewStreamsProps> = ({ project }) => {
  let cta: CallToAction | undefined;
  if (!project.streams?.length) {
    cta = {
      message: <>We didn't find any streams in <strong>{project.organization.name}/{project.name}</strong></>
    };
    if (project.permissions.create) {
      cta.buttons = [{ label: "Create stream", href: "/-/create/stream" }];
    }
  }

  return (
    <ContentContainer callToAction={cta}>
      <List>
        {project.streams.map(({ streamID, name, description }) => (
          <ListItem
            key={streamID}
            component={NakedLink}
            href={`/stream?organization_name=${toURLName(project.organization.name)}&project_name=${toURLName(
              project.name
            )}&stream_name=${toURLName(name)}`}
            as={`/${toURLName(project.organization.name)}/${toURLName(project.name)}/${toURLName(name)}`}
            button
            disableGutters
          >
            <ListItemAvatar>
              <Avatar size="list" label={name} />
            </ListItemAvatar>
            <ListItemText primary={toURLName(name)} secondary={description} />
          </ListItem>
        ))}
      </List>
    </ContentContainer>
  );
};

export default ViewStreams;
