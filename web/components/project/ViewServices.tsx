import {
  List,
  ListItem,
  ListItemAvatar,
  ListItemText,
  makeStyles,
} from "@material-ui/core";
import React, { FC } from "react";

import { ProjectByOrganizationAndName_projectByOrganizationAndName } from "apollo/types/ProjectByOrganizationAndName";
import Avatar from "components/Avatar";
import { NakedLink } from "components/Link";
import ContentContainer from "components/ContentContainer";
import { toURLName } from "lib/names";

const useStyles = makeStyles((theme) => ({}));

export interface ViewServicesProps {
  project: ProjectByOrganizationAndName_projectByOrganizationAndName;
}

const ViewServices: FC<ViewServicesProps> = ({ project }) => {
  let cta;
  if (!project.services?.length) {
    cta = {
      message: `${project.displayName || project.name} doesn't have any services`,
      buttons: [{ label: "Create service", href: "/-/create/service" }],
    };
  }
  const classes = useStyles();
  return (
    <ContentContainer paper maxWidth={"md"} callToAction={cta}>
      {!!project.services.length && (
        <List>
          {project.services.map(({ serviceID, name, description }) => (
            <ListItem
              key={serviceID}
              component={NakedLink}
              href={
                `/service?organization_name=${toURLName(project.organization.name)}` +
                `&project_name=${toURLName(project.name)}&service_name=${toURLName(name)}`
              }
              as={`/${toURLName(project.organization.name)}/${toURLName(project.name)}/-/services/${toURLName(name)}`}
              button
            >
              <ListItemAvatar>
                <Avatar size="list" label={name} />
              </ListItemAvatar>
              <ListItemText primary={toURLName(name)} secondary={description} />
            </ListItem>
          ))}
        </List>
      )}
    </ContentContainer>
  );
};

export default ViewServices;
