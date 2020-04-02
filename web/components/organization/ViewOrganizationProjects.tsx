import React, { FC } from "react";
import _ from "lodash";

import { List, ListItem, ListItemAvatar, ListItemText, makeStyles, Typography } from "@material-ui/core";

import { OrganizationByName_organizationByName, OrganizationByName_organizationByName_projects } from "../../apollo/types/OrganizationByName";
import { toURLName } from "../../lib/names";
import Avatar from "../Avatar";
import NextMuiLinkList from "../NextMuiLinkList";

const useStyles = makeStyles((theme) => ({
  noDataCaption: {
    color: theme.palette.text.secondary,
  },
}));

interface Props {
  organization: OrganizationByName_organizationByName;
}

const ViewOrganizationProjects: FC<Props> = ({ organization }) => {
  const classes = useStyles();
    
  return (
    <>
      <List>
        {organization.projects.map(({ projectID, name, displayName, description, photoURL }) => (
          <ListItem
            component={NextMuiLinkList}
            href={`/${toURLName(organization.name)}/${toURLName(name)}`}
            button
            disableGutters
            key={projectID}
          >
            <ListItemAvatar>
              <Avatar size="list" label={displayName || name} src={photoURL || undefined} />
            </ListItemAvatar>
            <ListItemText primary={displayName || name} secondary={description} />
          </ListItem>
        ))}
      </List>
      {organization.projects.length === 0 && (
        <Typography className={classes.noDataCaption} variant="body1" align="center">
          @{organization.name} doesn't have any projects to show
        </Typography>
      )}
    </>
  );
};

export default ViewOrganizationProjects;
