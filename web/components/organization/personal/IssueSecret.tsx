import { useMutation } from "@apollo/react-hooks";
import React, { FC } from "react";

import {
  Button,
  Card,
  CardContent,
  Dialog,
  DialogActions,
  DialogContent,
  DialogContentText,
  DialogTitle,
  Grid,
  makeStyles,
  TextField,
  Typography,
} from "@material-ui/core";

import { ISSUE_USER_SECRET, QUERY_USER_SECRETS } from "../../../apollo/queries/secret";
import { IssueUserSecret, IssueUserSecretVariables } from "../../../apollo/types/IssueUserSecret";

const useStyles = makeStyles((theme) => ({
  issueSecretButton: {
    display: "block",
    minHeight: theme.spacing(10),
    textAlign: "left",
    textTransform: "none",
  },
}));

interface IssueSecretProps {
  userID: string;
}

const IssueSecret: FC<IssueSecretProps> = ({ userID }) => {
  const [dialogOpen, setDialogOpen] = React.useState(false);
  const [readOnlySecret, setReadOnlySecret] = React.useState(true);
  const [newSecretString, setNewSecretString] = React.useState("");

  const [issueSecret, { loading, error }] = useMutation<IssueUserSecret, IssueUserSecretVariables>(ISSUE_USER_SECRET, {
    onCompleted: (data) => {
      setNewSecretString(data.issueUserSecret.token);
      closeDialog();
    },
  });

  const openDialog = (readonly: boolean) => {
    setReadOnlySecret(readonly);
    setDialogOpen(true);
  };

  const closeDialog = () => {
    setDialogOpen(false);
  };

  let input: any = null;
  const classes = useStyles();
  return (
    <div>
      <Grid container spacing={2}>
        <Grid item xs={12} md={6}>
          <Button
            color="primary"
            variant="outlined"
            fullWidth
            className={classes.issueSecretButton}
            onClick={() => openDialog(true)}
          >
            <Typography variant="button" display="block">
              Create new read-only secret
            </Typography>
            <Typography variant="caption" display="block">
              Grants read access to all streams you have access to.
            </Typography>
            <Typography variant="caption" display="block">
              <strong>Use to</strong> quickly get started with integrating data from Beneath
            </Typography>
          </Button>
        </Grid>
        <Grid item xs={12} md={6}>
          <Button
            color="secondary"
            variant="outlined"
            fullWidth
            className={classes.issueSecretButton}
            onClick={() => openDialog(false)}
          >
            <Typography variant="button" display="block">
              Create new command-line secret
            </Typography>
            <Typography variant="caption" display="block">
              Grants full access to modify resources
            </Typography>
            <Typography variant="caption" display="block">
              <strong>Use to</strong> connect from the Beneath command-line app
            </Typography>
          </Button>
        </Grid>
        {newSecretString !== "" && (
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
      <Dialog open={dialogOpen} onClose={closeDialog} aria-labelledby="form-dialog-title" fullWidth>
        <form onSubmit={(e) => e.preventDefault()}>
          <DialogTitle id="form-dialog-title">
            Issue {readOnlySecret ? "read-only" : "command-line"} secret
          </DialogTitle>
          <DialogContent>
            <TextField
              autoFocus
              fullWidth
              margin="dense"
              id="name"
              label="Enter a description of the secret"
              defaultValue={`My ${readOnlySecret ? "read-only" : "command-line"} secret`}
              inputRef={(node) => (input = node)}
            />
          </DialogContent>
          {error && (
            <DialogContent>
              <DialogContentText>An error occurred: {JSON.stringify(error)}</DialogContentText>
            </DialogContent>
          )}
          <DialogActions>
            <Button color="primary" onClick={closeDialog}>
              Cancel
            </Button>
            <Button
              type="submit"
              color="primary"
              disabled={loading}
              onClick={() => {
                issueSecret({
                  variables: {
                    description: input.value,
                    publicOnly: readOnlySecret,
                    readOnly: readOnlySecret,
                  },
                  update: (cache, { data }) => {
                    if (data) {
                      const queryData = cache.readQuery({
                        query: QUERY_USER_SECRETS,
                        variables: { userID },
                      }) as any;
                      cache.writeQuery({
                        query: QUERY_USER_SECRETS,
                        variables: { userID },
                        data: { secretsForUser: queryData.secretsForUser.concat([data.issueUserSecret.secret]) },
                      });
                    }
                  },
                });
              }}
            >
              Issue secret
            </Button>
          </DialogActions>
        </form>
      </Dialog>
    </div>
  );
};

export default IssueSecret;
