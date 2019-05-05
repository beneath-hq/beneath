import React, { Component } from "react";
import gql from "graphql-tag";
import { Query, Mutation } from "react-apollo";

import Button from "@material-ui/core/Button";
import Card from "@material-ui/core/Card";
import CardContent from "@material-ui/core/CardContent";
import Container from "@material-ui/core/Container";
import DeleteIcon from "@material-ui/icons/Delete";
import Dialog from "@material-ui/core/Dialog";
import DialogTitle from "@material-ui/core/DialogTitle";
import DialogContent from "@material-ui/core/DialogContent";
import DialogContentText from "@material-ui/core/DialogContentText";
import DialogActions from "@material-ui/core/DialogActions";
import Grid from "@material-ui/core/Grid";
import IconButton from "@material-ui/core/IconButton";
import List from "@material-ui/core/List";
import ListItem from "@material-ui/core/ListItem";
import ListItemIcon from "@material-ui/core/ListItemIcon";
import ListItemText from "@material-ui/core/ListItemText";
import ListItemSecondaryAction from "@material-ui/core/ListItemSecondaryAction";
import Moment from "react-moment";
import TextField from "@material-ui/core/TextField";
import Typography from "@material-ui/core/Typography";
import { makeStyles } from "@material-ui/core";

import Loading from "../components/Loading";
import Page from "../components/Page";
import { AuthRequired } from "../hocs/auth";

// TODO
/*
  Section 1: Edit name and bio
*/

const QUERY_ME = gql`
  query {
    me {
      userId
      email
      username
      name
      bio
      photoUrl
      createdOn
      updatedOn
      keys {
        keyId
        description
        prefix
        role
        createdOn
      }
    }
  }
`;

const ISSUE_KEY = gql`
  mutation IssueKey($description: String!, $readonly: Boolean!) {
    issueKey(description: $description, readonly: $readonly) {
      keyString
      key {
        keyId
        description
        prefix
        role
        createdOn
      }
    }
  }
`;

const REVOKE_KEY = gql`
  mutation RevokeKey($keyId: ID!) {
    revokeKey(keyId: $keyId)
  }
`;

const useStyles = makeStyles((theme) => ({
  profileContent: {
    padding: theme.spacing(8, 0, 6),
  },
  section: {
    paddingBottom: theme.spacing(6),
  },
  issueKeyButton: {
    display: "block",
    minHeight: theme.spacing(10),
    textAlign: "left",
    textTransform: "none",
  },
}));

export default () => {
  const classes = useStyles();
  return (
    <AuthRequired>
      <Page title="Profile">
        <div className={classes.profileContent}>
          <Container maxWidth="md">
            <Query query={QUERY_ME}>
              {({ loading, error, data }) => {
                if (loading) return <Loading justify="center" />;
                if (error) return <p>Error: {JSON.stringify(error)}</p>;
                let { me } = data;
                return (
                  <React.Fragment>
                    <div className={classes.section}>
                      <Button size="large" color="primary" variant="outlined" fullWidth href="/auth/logout">
                        Logout
                      </Button>
                    </div>
                    <EditProfile me={me} />
                    <ManageKeys me={me} />
                  </React.Fragment>
                );
              }}
            </Query>
          </Container>
        </div>
      </Page>
    </AuthRequired>
  );
};

const EditProfile = ({ me }) => {
  const classes = useStyles();
  return (
    // <Mutation mutation={MUTATE_ME} key={me.userId}>
    //   {(updateMe) => (
        <div className={classes.section}>
          <Typography component="h3" variant="h4" gutterBottom>
            Edit profile
          </Typography>
          <Typography variant="subtitle1" gutterBottom>
            You signed up <Moment fromNow date={me.createdOn} /> and
            last updated your profile <Moment fromNow date={me.updatedOn} />.
          </Typography>
          <form onSubmit={(e) => {
              e.preventDefault();
              // username, name, bio
              // updateMe({ variables: {
              //   id, type: input.value
              // } });
              // input.value = "";
            }}
          >
          </form>
        </div>
    //   )}
    // </Mutation>
  );
};

const ManageKeys = ({ me }) => {
  const classes = useStyles();
  return (
    <div className={classes.section}>
      <Typography component="h3" variant="h4" gutterBottom>
        Manage keys
      </Typography>
      <IssueKey me={me} />
      <ViewKeys me={me} />
    </div>
  );
};

const IssueKey = ({ me }) => {
  const [dialogOpen, setDialogOpen] = React.useState(false);
  const [readonlyKey, setReadonlyKey] = React.useState(false);
  const [newKeyString, setNewKeyString] = React.useState(null);
  
  const openDialog = (readonly) => {
    setReadonlyKey(readonly);
    setDialogOpen(true);
  };
  const closeDialog = (data) => {
    setDialogOpen(false);
  };

  let input = null;

  const classes = useStyles();

  return (
    <div>
      <Grid container spacing={2}>
        <Grid item xs={12} md={6}>
          <Button color="secondary" variant="outlined" fullWidth
            className={classes.issueKeyButton} onClick={() => openDialog(false)}>
            <Typography variant="button" display="block">
              Issue read/write key
            </Typography>
            <Typography variant="caption" display="block">
              Grants access to read and mutate data. E.g. pushing external data into Beneath.
            </Typography>
          </Button>
        </Grid>
        <Grid item xs={12} md={6}>
          <Button color="primary" variant="outlined" fullWidth
            className={classes.issueKeyButton} onClick={() => openDialog(true)}>
            <Typography variant="button" display="block">
              Issue read-only key
            </Typography>
            <Typography variant="caption" display="block">
              Grants access to read data. E.g. reading Beneath data from an external application.
            </Typography>
          </Button>
        </Grid>
        {newKeyString && (
          <Grid item xs={12}>
            <Card>
              <CardContent>
                <Typography variant="h6" color="textSecondary" gutterBottom>
                  Here is your new key:
                </Typography>
                <Typography color="textSecondary" noWrap gutterBottom>
                  {newKeyString}
                </Typography>
                <Typography variant="body2">
                  The key will only be shown this once – remember to keep it safe!
                </Typography>
              </CardContent>
            </Card>
          </Grid>
        )}
      </Grid>
      <Mutation mutation={ISSUE_KEY}
        onCompleted={({ issueKey }) => {
          setNewKeyString(issueKey.keyString);
          closeDialog();
        }}
        update={(cache, { data: { issueKey } }) => {
          const { me } = cache.readQuery({ query: QUERY_ME });
          cache.writeQuery({
            query: QUERY_ME,
            data: { me: { ...me, keys: me.keys.concat([issueKey.key]) } },
          });
        }}
      >
        {(issueKey, { loading, error }) => (
          <Dialog open={dialogOpen} onClose={closeDialog} aria-labelledby="form-dialog-title" fullWidth>
            <DialogTitle id="form-dialog-title">Issue key</DialogTitle>
            <DialogContent>
              <DialogContentText>
                Enter a description for the key
              </DialogContentText>
              <TextField autoFocus margin="dense" id="name" label="Key Description" fullWidth inputRef={(node) => input = node} />
            </DialogContent>
            <DialogActions>
              <Button color="primary" onClick={closeDialog}>
                Cancel
              </Button>
              <Button color="primary" disabled={loading} error={error} onClick={() => {
                issueKey({ variables: { description: input.value, readonly: readonlyKey } });
              }}>
                Issue key
              </Button>
            </DialogActions>
          </Dialog>
        )}
      </Mutation>
    </div>
  );
};

const prettyRoles = {
  "personal": "Browser login",
  "readonly": "Read-only",
  "readwrite": "Read/write",
};

const ViewKeys = ({ me }) => {
  return (
    <List dense={true}>
      { me.keys.map(({ createdOn, description, keyId, prefix, role }) => (
        <ListItem key={keyId} disableGutters>
          <ListItemText
            primary={
              <React.Fragment>
                {description || <emph>No description</emph>}
                {" "}(starts with <strong>{prefix}</strong>)
              </React.Fragment>
            }
            secondary={
              <React.Fragment>
                {prettyRoles[role]} key issued <Moment fromNow date={createdOn} />
              </React.Fragment>
            }
          />
          <ListItemSecondaryAction>
            <Mutation mutation={REVOKE_KEY} update={(cache, { data: { revokeKey } }) => {
              if (revokeKey) {
                const { me } = cache.readQuery({ query: QUERY_ME });
                cache.writeQuery({
                  query: QUERY_ME,
                  data: { me: { ...me, keys: me.keys.filter((key) => key.keyId !== keyId) } },
                });
              }
            }}>
              {(revokeKey, { loading, error }) => (
                <IconButton edge="end" aria-label="Delete" onClick={() => {
                  revokeKey({ variables: { keyId } });
                }}>
                  {loading ? <Loading size={20} /> : <DeleteIcon />}
                </IconButton>
              )}
            </Mutation>
          </ListItemSecondaryAction>
        </ListItem>
      ))}
    </List>
  );
};
