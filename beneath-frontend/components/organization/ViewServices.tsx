import React, { FC } from "react";

import { List, ListItem, ListItemText, makeStyles, Typography } from "@material-ui/core";

import { OrganizationByName_organizationByName } from "../../apollo/types/OrganizationByName";
import NextMuiLinkList from "../NextMuiLinkList";

const useStyles = makeStyles((theme) => ({
  noDataCaption: {
    color: theme.palette.text.secondary,
  },
}));

interface Props {
  organization: OrganizationByName_organizationByName;
}

const ViewServices: FC<Props> = ({ organization }) => {
  const classes = useStyles();
  return (
    <>
      <List>
        {organization.services.map(({ serviceID, name }) => (
          <ListItem
            component={NextMuiLinkList}
            href={`/`}
            button
            disableGutters
            key={serviceID}
            as={`/`}
          >
            <ListItemText primary={ name } secondary={name} />
          </ListItem>
        ))}
      </List>
      {organization.services.length === 0 && (
        <Typography className={classes.noDataCaption} variant="body1" align="center">
          {organization.name} doesn't have any services
        </Typography>
      )}
    </>
  );
};

export default ViewServices;
