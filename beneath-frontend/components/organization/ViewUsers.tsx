import React, { FC } from "react";
import { useQuery, useMutation } from "@apollo/react-hooks"
import { List, ListItem, ListItemAvatar, ListItemText, makeStyles, Typography, TextField, Button, Grid } from "@material-ui/core";

import useMe from "../../hooks/useMe";
import { OrganizationByName_organizationByName } from "../../apollo/types/OrganizationByName";
import { UsersOrganizationPermissions, UsersOrganizationPermissionsVariables } from "../../apollo/types/UsersOrganizationPermissions"
import { InviteUserToOrganization, InviteUserToOrganizationVariables } from "../../apollo/types/InviteUserToOrganization";
import { QUERY_USERS_ORGANIZATION_PERMISSIONS, ADD_USER_TO_ORGANIZATION } from "../../apollo/queries/organization"

import NextMuiLinkList from "../NextMuiLinkList";
import Avatar from "../Avatar";
import Loading from "../Loading"
import { toURLName } from "../../lib/names";

const useStyles = makeStyles((theme) => ({
  noDataCaption: {
    color: theme.palette.text.secondary,
  },
}));

interface Props {
  organization: OrganizationByName_organizationByName;
}

// TODO: show admin tag next to every user
// TODO: show quotas for every user (from organization.users.readquota)
const ViewUsers: FC<Props> = ({ organization }) => {
  const [username, setUsername] = React.useState("")
  const [admin, setAdmin] = React.useState(false)
  const [error, setError] = React.useState("")
  const [isInviteSent, setIsInviteSent] = React.useState(false)

  const classes = useStyles();
  const me = useMe();

  const { loading: queryLoading, error: queryError, data } = useQuery<UsersOrganizationPermissions, UsersOrganizationPermissionsVariables>(QUERY_USERS_ORGANIZATION_PERMISSIONS, {
    variables: {
      organizationID: organization.organizationID,
    },});
  const [inviteUser] = useMutation<InviteUserToOrganization, InviteUserToOrganizationVariables>(ADD_USER_TO_ORGANIZATION, {
    onCompleted: ({ inviteUserToOrganization }) => {
      if (inviteUserToOrganization) {
        setError("")
        setIsInviteSent(true)
      }
    },
    onError: ( error ) => {
      setError(error.message.replace("GraphQL error:", ""))
      setIsInviteSent(false)
    },
  })

  if (queryLoading) return <Loading />
  if (queryError || !data) return <p>Error: {JSON.stringify(queryError)}</p>

  const { usersOrganizationPermissions } = data;

  // check to see if user is an organization admin
  if (me) {
    const meOrganizationPermissions = usersOrganizationPermissions.find(x => x.user.userID == me.userID)
    if (meOrganizationPermissions && meOrganizationPermissions.admin && !admin) {
      setAdmin(true)
    } 
  }
  
  const handleChange = (event: any) => {
    if(isInviteSent) {
      setIsInviteSent(false)
    }
    setUsername(event.target.value)
  }
  
  const handleInviteUserFormSubmit = (ev: any) => {
    // We don't want to let default form submission happen here, which would refresh the page.
    ev.preventDefault()
    inviteUser({
      variables: {
        username: username,
        organizationID: organization.organizationID,
        view: true,
        admin: false,
      },
    })
  }

  return (
    <>
      <Grid container direction="column" spacing={4}>
        <Grid item>
          <List>
            {organization.users.map(({ userID, name, username, photoURL, readQuota, writeQuota }) => (
              <ListItem
                component={NextMuiLinkList}
                href={`/user?name=${toURLName(username)}`}
                button
                disableGutters
                key={userID}
                as={`/users/${toURLName(username)}`}
              >
                <ListItemAvatar>
                  <Avatar size="list" label={username || name} src={photoURL || undefined} />
                </ListItemAvatar>
                <ListItemText primary={username || name} secondary={name} />
              </ListItem>
            ))}
          </List>
        </Grid>
        {admin && (
        <Grid item>
          <form onSubmit={handleInviteUserFormSubmit}>
            <Grid container alignItems="center" spacing={2}>
              <Grid item>
                <TextField
                  required
                  id="username"
                  name="username"
                  label="Username"
                  value={username}
                  onChange={handleChange}
                />
              </Grid>
              <Grid item>
                <Button type="submit" variant="contained" color="primary">
                  Invite User
                </Button>
              </Grid>
            </Grid>
          </form>
          {error && (
            <Typography variant="caption" color="error">{error}</Typography>
          )}
          {isInviteSent && (
            <Typography variant="caption">An invitation to join your organization has been sent to {username}'s email</Typography>
          )}
        </Grid>
        )}
      </Grid>
      {organization.users.length === 0 && (
        <Typography className={classes.noDataCaption} variant="body1" align="center">
          {organization.name} doesn't have any users... that's strange
        </Typography>
      )}
    </>
  );
};

export default ViewUsers;
