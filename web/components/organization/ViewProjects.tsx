import React, { FC } from "react";

import { List, ListItem, ListItemAvatar, ListItemText, makeStyles, Typography } from "@material-ui/core";

import { OrganizationByName_organizationByName } from "../../apollo/types/OrganizationByName";
import { toURLName } from "../../lib/names";
import Avatar from "../Avatar";
import NextMuiLinkList from "../NextMuiLinkList";

const useStyles = makeStyles((theme) => ({
  noDataCaption: {
    color: theme.palette.text.secondary,
  },
}));

export interface ViewProjectsProps {
  organization: OrganizationByName_organizationByName;
}

const ViewProjects: FC<ViewProjectsProps> = ({ organization }) => {
  const classes = useStyles();
  return (
    <>
      <List>
        {organization.projects.map(({ projectID, name, displayName, description, photoURL }) => (
          <ListItem
            key={projectID}
            component={NextMuiLinkList}
            href={`/project?organization_name=${toURLName(organization.name)}&project_name=${toURLName(name)}`}
            as={`/${toURLName(organization.name)}/${toURLName(name)}`}
            button
            disableGutters
          >
            <ListItemAvatar>
              <Avatar size="list" label={displayName || name} src={photoURL} />
            </ListItemAvatar>
            <ListItemText primary={displayName || name} secondary={description} />
          </ListItem>
        ))}
      </List>
      {organization.projects.length === 0 && (
        <Typography className={classes.noDataCaption} variant="body1" align="center">
          {organization.displayName} doesn't have any projects to show
        </Typography>
      )}
    </>
  );
};

export default ViewProjects;
