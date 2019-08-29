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

import {
  QUERY_USER_SECRETS,
  QUERY_PROJECT_SECRETS,
  ISSUE_USER_SECRET,
  ISSUE_PROJECT_SECRET
} from "../../apollo/queries/secret";

const useStyles = makeStyles((theme) => ({
  issueSecretButton: {
    display: "block",
    minHeight: theme.spacing(10),
    textAlign: "left",
    textTransform: "none",
  },
}));

// entity is either user or project
const IssueSecret = ({ entityName, entityID }) => {
  const [dialogOpen, setDialogOpen] = React.useState(false);
  const [readonlySecret, setReadonlySecret] = React.useState(false);
  const [newSecretString, setNewSecretString] = React.useState(null);

  const openDialog = (readonly) => {
    setReadonlySecret(readonly);
    setDialogOpen(true);
  };
  const closeDialog = (data) => {
    setDialogOpen(false);
  };

  let query;
  let queryKey;
  let mutation;
  let mutationKey;
  let entityIDKey;
  if (entityName === "user") {
    query = QUERY_USER_SECRETS;
    queryKey = "secretsForUser";
    mutation = ISSUE_USER_SECRET;
    mutationKey = "issueUserSecret";
    entityIDKey = "userID";
  } else if (entityName === "project") {
    query = QUERY_PROJECT_SECRETS;
    queryKey = "secretsForProject";
    mutation = ISSUE_PROJECT_SECRET;
    mutationKey = "issueProjectSecret";
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
            className={classes.issueSecretButton} onClick={() => openDialog(false)}>
            <Typography variant="button" display="block">
              Issue read/write secret
            </Typography>
            <Typography variant="caption" display="block">
              Grants access to read and mutate data. E.g. pushing external data into Beneath.
            </Typography>
          </Button>
        </Grid>
        <Grid item xs={12} md={6}>
          <Button color="primary" variant="outlined" fullWidth
            className={classes.issueSecretButton} onClick={() => openDialog(true)}>
            <Typography variant="button" display="block">
              Issue read-only secret
            </Typography>
            <Typography variant="caption" display="block">
              Grants access to read data. E.g. reading Beneath data from an external application.
            </Typography>
          </Button>
        </Grid>
        {newSecretString && (
          <Grid item xs={12}>
            <Card>
              <CardContent>
                <Typography variant="h6" color="textSecondary" gutterBottom>
                  Here is your new secret:
                </Typography>
                <Typography color="textSecondary" noWrap gutterBottom>
                  {newSecretString}
                </Typography>
                <Typography variant="body2">
                  The secret will only be shown this once – remember to keep it safe!
                </Typography>
              </CardContent>
            </Card>
          </Grid>
        )}
      </Grid>
      <Mutation mutation={mutation}
        onCompleted={(data) => {
          setNewSecretString(data[mutationKey].secretString);
          closeDialog();
        }}
        update={(cache, { data }) => {
          const queryData = cache.readQuery({ query: query, variables: { [entityIDKey]: entityID } });
          cache.writeQuery({
            query: query,
            variables: { [entityIDKey]: entityID },
            data: { [queryKey]: queryData[queryKey].concat([data[mutationKey].secret]) },
          });
        }}
      >
        {(issueSecret, { loading, error }) => (
          <Dialog open={dialogOpen} onClose={closeDialog} aria-labelledby="form-dialog-title" fullWidth>
            <DialogTitle id="form-dialog-title">Issue secret</DialogTitle>
            <DialogContent>
              <DialogContentText>
                Enter a description for the secret
              </DialogContentText>
              <TextField autoFocus margin="dense" id="name" label="Secret Description" fullWidth inputRef={(node) => input = node} />
            </DialogContent>
            <DialogActions>
              <Button color="primary" onClick={closeDialog}>
                Cancel
              </Button>
              <Button color="primary" disabled={loading} error={error} onClick={() => {
                issueSecret({
                  variables: { [entityIDKey]: entityID, description: input.value, readonly: readonlySecret },
                });
              }}>
                Issue secret
              </Button>
            </DialogActions>
          </Dialog>
        )}
      </Mutation>
    </div>
  );
};

export default IssueSecret;
