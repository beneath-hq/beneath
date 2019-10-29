import React, { FC } from "react";
import { Mutation } from "react-apollo";

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

import { ISSUE_USER_SECRET, QUERY_USER_SECRETS } from "../../apollo/queries/secret";
import { IssueUserSecret, IssueUserSecretVariables } from "../../apollo/types/IssueUserSecret";

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
  const [newSecretString, setNewSecretString] = React.useState("");

  const openDialog = () => {
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
        <Grid item xs={12} md={12}>
          <Button
            color="secondary"
            variant="outlined"
            fullWidth
            className={classes.issueSecretButton}
            onClick={() => openDialog()}
          >
            <Typography variant="button" display="block">
              Issue secret
            </Typography>
            <Typography variant="caption" display="block">
              Grants full access to read data and create/update/delete new models and streams
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
      <Mutation<IssueUserSecret, IssueUserSecretVariables>
        mutation={ISSUE_USER_SECRET}
        onCompleted={(data) => {
          setNewSecretString(data.issueUserSecret.token);
          closeDialog();
        }}
        update={(cache, { data }) => {
          if (data) {
            const queryData = cache.readQuery({ query: QUERY_USER_SECRETS, variables: { userID } }) as any;
            cache.writeQuery({
              query: QUERY_USER_SECRETS,
              variables: { userID },
              data: { secretsForUser: queryData.secretsForUser.concat([data.issueUserSecret.secret]) },
            });
          }
        }}
      >
        {(issueSecret, { loading, error }) => (
          <Dialog open={dialogOpen} onClose={closeDialog} aria-labelledby="form-dialog-title" fullWidth>
            <form onSubmit={(e) => e.preventDefault()}>
              <DialogTitle id="form-dialog-title">Issue secret</DialogTitle>
              <DialogContent>
                <DialogContentText>Enter a description for the secret</DialogContentText>
                <TextField
                  autoFocus
                  margin="dense"
                  id="name"
                  label="Secret Description"
                  fullWidth
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
                        publicOnly: false,
                        readOnly: false,
                      },
                    });
                  }}
                >
                  Issue secret
                </Button>
              </DialogActions>
            </form>
          </Dialog>
        )}
      </Mutation>
    </div>
  );
};

export default IssueSecret;
