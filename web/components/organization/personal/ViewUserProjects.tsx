import React, { FC } from "react";

import { List, ListItem, ListItemAvatar, ListItemText, makeStyles, Typography } from "@material-ui/core";

import { User_user } from "../../../apollo/types/User";
import { toURLName } from "../../../lib/names";
import Avatar from "../../Avatar";
import NextMuiLinkList from "../../NextMuiLinkList";

const useStyles = makeStyles((theme) => ({
  noDataCaption: {
    color: theme.palette.text.secondary,
  },
}));

interface Props {
  user: User_user;
}

const ViewUserProjects: FC<Props> = ({ user }) => {
  const classes = useStyles();
  return (
    <>
      <List>
        {user.projects.map(({ projectID, name, displayName, description, photoURL, organization }) => (
          <ListItem
            component={NextMuiLinkList}
            href={`/project?organization_name=${toURLName(organization.name)}&project_name=${toURLName(name)}`}
            as={`/${toURLName(organization.name)}/${toURLName(name)}`}
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
      {user.projects.length === 0 && (
        <Typography className={classes.noDataCaption} variant="body1" align="center">
          @{user.username} isn't a member of any projects
        </Typography>
      )}
    </>
  );
};

export default ViewUserProjects;
