import React, { FC } from "react";

import { List, ListItem, ListItemAvatar, ListItemText, makeStyles, Typography } from "@material-ui/core";

import { OrganizationByName_organizationByName_PrivateOrganization } from "../../apollo/types/OrganizationByName";
import { toURLName } from "../../lib/names";
import Avatar from "../Avatar";
import NextMuiLinkList from "../NextMuiLinkList";

const useStyles = makeStyles((theme) => ({
  noDataCaption: {
    color: theme.palette.text.secondary,
  },
}));

export interface ViewProjectsProps {
  organization: OrganizationByName_organizationByName_PrivateOrganization;
}

const ViewProjects: FC<ViewProjectsProps> = ({ organization }) => {
  const classes = useStyles();
  return (
    <>
      <List>
        {organization.services.map(({ serviceID, name, kind }) => (
          <ListItem
            key={serviceID}
            // component={NextMuiLinkList}
            // href={`/-/service?organization_name=${toURLName(organization.name)}&service_name=${toURLName(name)}`}
            // as={`/${toURLName(organization.name)}/-/services/${toURLName(name)}`}
            button
            disableGutters
          >
            <ListItemAvatar>
              <Avatar size="list" label={name} />
            </ListItemAvatar>
            <ListItemText primary={name} secondary={kind && kind[0].toUpperCase() + kind.slice(1)} />
          </ListItem>
        ))}
      </List>
      {organization.services.length === 0 && (
        <Typography className={classes.noDataCaption} variant="body1" align="center">
          {organization.displayName} doesn't have any services
        </Typography>
      )}
    </>
  );
};

export default ViewProjects;
