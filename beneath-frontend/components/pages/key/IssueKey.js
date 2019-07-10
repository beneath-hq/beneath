import React from "react";
import { Mutation } from "react-apollo";

import Button from "@material-ui/core/Button";
import Card from "@material-ui/core/Card";
import CardContent from "@material-ui/core/CardContent";
import Dialog from "@material-ui/core/Dialog";
import DialogTitle from "@material-ui/core/DialogTitle";
import DialogContent from "@material-ui/core/DialogContent";
import DialogContentText from "@material-ui/core/DialogContentText";
import DialogActions from "@material-ui/core/DialogActions";
import Grid from "@material-ui/core/Grid";
import TextField from "@material-ui/core/TextField";
import Typography from "@material-ui/core/Typography";
import { makeStyles } from "@material-ui/core";

import { QUERY_USER_KEYS, QUERY_PROJECT_KEYS, ISSUE_USER_KEY, ISSUE_PROJECT_KEY } from "../../../queries/key";

const useStyles = makeStyles((theme) => ({
  issueKeyButton: {
    display: "block",
    minHeight: theme.spacing(10),
    textAlign: "left",
    textTransform: "none",
  },
}));

// entity is either user or project
const IssueKey = ({ entityName, entityID }) => {
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

  let query, queryKey, mutation, mutationKey, entityIDKey;
  if (entityName === "user") {
    query = QUERY_USER_KEYS;
    queryKey = "keysForUser";
    mutation = ISSUE_USER_KEY;
    mutationKey = "issueUserKey";
    entityIDKey = "userID";
  } else if (entityName === "project") {
    query = QUERY_PROJECT_KEYS;
    queryKey = "keysForProject";
    mutation = ISSUE_PROJECT_KEY;
    mutationKey = "issueProjectKey";
    entityIDKey = "projectID";
  } else {
    console.error("expected entity to be 'user' or 'project'")
  }


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
      <Mutation mutation={mutation}
        onCompleted={(data) => {
          setNewKeyString(data[mutationKey].keyString);
          closeDialog();
        }}
        update={(cache, { data }) => {
          const queryData = cache.readQuery({ query: query, variables: { [entityIDKey]: entityID } });
          cache.writeQuery({
            query: query,
            variables: { [entityIDKey]: entityID },
            data: { [queryKey]: queryData[queryKey].concat([data[mutationKey].key]) },
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
                issueKey({ variables: { [entityIDKey]: entityID, description: input.value, readonly: readonlyKey } });
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

export default IssueKey;
