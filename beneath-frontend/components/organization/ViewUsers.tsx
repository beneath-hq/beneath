import React, { FC } from "react";

import { List, ListItem, ListItemAvatar, ListItemText, makeStyles, Typography } from "@material-ui/core";

import { OrganizationByName_organizationByName } from "../../apollo/types/OrganizationByName";
import NextMuiLinkList from "../NextMuiLinkList";
import { toURLName } from "../../lib/names";
import Avatar from "../Avatar";

const useStyles = makeStyles((theme) => ({
  noDataCaption: {
    color: theme.palette.text.secondary,
  },
}));

interface Props {
  organization: OrganizationByName_organizationByName;
}

const ViewUsers: FC<Props> = ({ organization }) => {
  const classes = useStyles();
  return (
    <>
      <List>
        {organization.users.map(({ userID, name, username, photoURL }) => (
          <ListItem
            component={NextMuiLinkList}
            href={`/user?name=${toURLName(name)}`}
            button
            disableGutters
            key={userID}
            as={`/users/${toURLName(name)}`}
          >
            <ListItemAvatar>
              <Avatar size="list" label={username || name} src={photoURL || undefined} />
            </ListItemAvatar>
            <ListItemText primary={username || name} secondary={name} />
          </ListItem>
        ))}
      </List>
      {organization.users.length === 0 && (
        <Typography className={classes.noDataCaption} variant="body1" align="center">
          {organization.name} doesn't have any users... that's strange
        </Typography>
      )}
    </>
  );
};

export default ViewUsers;
