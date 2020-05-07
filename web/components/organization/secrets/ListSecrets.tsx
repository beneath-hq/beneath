import { useMutation, useQuery } from "@apollo/react-hooks";
import React, { FC } from "react";
import Moment from "react-moment";

import {
  Button,
  Dialog,
  DialogActions,
  DialogContent,
  DialogContentText,
  DialogTitle,
  IconButton,
  List,
  ListItem,
  ListItemSecondaryAction,
  ListItemText,
} from "@material-ui/core";
import DeleteIcon from "@material-ui/icons/Delete";

import { QUERY_USER_SECRETS, REVOKE_USER_SECRET } from "../../../apollo/queries/secret";
import { RevokeUserSecret, RevokeUserSecretVariables } from "../../../apollo/types/RevokeUserSecret";
import { SecretsForUser, SecretsForUserVariables } from "../../../apollo/types/SecretsForUser";
import Loading from "../../Loading";

export interface ListSecretsProps {
  userID: string;
}

const ListSecrets: FC<ListSecretsProps> = ({ userID }) => {
  const [openDialogue, setOpenDialogue] = React.useState(false);

  const { loading, error, data } = useQuery<SecretsForUser, SecretsForUserVariables>(QUERY_USER_SECRETS, {
    variables: { userID },
  });

  const [revokeSecret, { loading: mutLoading }] = useMutation<RevokeUserSecret, RevokeUserSecretVariables>(
    REVOKE_USER_SECRET
  );

  if (loading) {
    return <Loading justify="center" />;
  }

  if (error || !data) {
    return <p>Error: {JSON.stringify(error)}</p>;
  }

  return (
    <List dense={true}>
      {data.secretsForUser
        .filter((secret) => secret.description !== "Browser session")
        .map(({ createdOn, description, userSecretID, prefix }) => (
          <ListItem key={userSecretID} disableGutters>
            <ListItemText
              primary={
                <React.Fragment>
                  {description || <em>No description</em>} (starts with <strong>{prefix}</strong>)
                </React.Fragment>
              }
              secondary={
                <React.Fragment>
                  Secret issued <Moment fromNow date={createdOn} />
                </React.Fragment>
              }
            />
            <ListItemSecondaryAction>
              <IconButton
                edge="end"
                aria-label="Delete"
                disabled={mutLoading}
                onClick={() => {
                  setOpenDialogue(true);
                }}
              >
                <DeleteIcon />
              </IconButton>
              <Dialog open={openDialogue}>
                <DialogTitle id="alert-dialog-title">{"Are you sure you want to delete this secret?"}</DialogTitle>
                <DialogContent>
                  <DialogContentText id="alert-dialog-description">
                    Any environments (your CLI, any services, Jupyter notebooks, etc.) that rely on this secret will no
                    longer work. For the affected environments, you'll have to issue a new secret and re-authenticate.
                  </DialogContentText>
                </DialogContent>
                <DialogActions>
                  <Button
                    color="primary"
                    autoFocus
                    onClick={() => {
                      setOpenDialogue(false);
                    }}
                  >
                    No, go back
                  </Button>
                  <Button
                    color="primary"
                    autoFocus
                    onClick={() => {
                      revokeSecret({
                        variables: { secretID: userSecretID },
                        update: (cache, { data }) => {
                          if (data && data.revokeUserSecret) {
                            const queryData = cache.readQuery({
                              query: QUERY_USER_SECRETS,
                              variables: { userID },
                            }) as any;
                            const filtered = queryData.secretsForUser.filter(
                              (secret: any) => secret.userSecretID !== userSecretID
                            );
                            cache.writeQuery({
                              query: QUERY_USER_SECRETS,
                              variables: { userID },
                              data: { secretsForUser: filtered },
                            });
                          }
                        },
                      });
                    }}
                  >
                    Yes, I'm sure
                  </Button>
                </DialogActions>
              </Dialog>
            </ListItemSecondaryAction>
          </ListItem>
        ))}
    </List>
  );
};

export default ListSecrets;
