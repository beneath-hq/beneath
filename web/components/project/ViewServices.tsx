import React, { FC } from "react";

import { List, ListItem, ListItemAvatar, ListItemText, makeStyles, Typography } from "@material-ui/core";

import { ProjectByOrganizationAndName_projectByOrganizationAndName } from "../../apollo/types/ProjectByOrganizationAndName";
import { toURLName } from "../../lib/names";
import Avatar from "../Avatar";
import { NakedLink } from "../Link";

const useStyles = makeStyles((theme) => ({
  noDataCaption: {
    color: theme.palette.text.secondary,
  },
}));

export interface ViewServicesProps {
  project: ProjectByOrganizationAndName_projectByOrganizationAndName;
}

const ViewServices: FC<ViewServicesProps> = ({ project }) => {
  const classes = useStyles();
  return (
    <>
      <List>
        {project.services.map(({ serviceID, name, description }) => (
          <ListItem
            key={serviceID}
            component={NakedLink}
            href={`/service?organization_name=${toURLName(project.organization.name)}&project_name=${toURLName(
              project.name
            )}&service_name=${toURLName(name)}`}
            as={`/${toURLName(project.organization.name)}/${toURLName(project.name)}/-/services/${toURLName(name)}`}
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
      {project.services.length === 0 && (
        <Typography className={classes.noDataCaption} variant="body1" align="center">
          {project.displayName} doesn't have any services
        </Typography>
      )}
    </>
  );
};

export default ViewServices;
