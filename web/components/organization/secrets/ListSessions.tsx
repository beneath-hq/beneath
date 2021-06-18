import { useMutation, useQuery } from "@apollo/client";
import {
  Button,
  Dialog,
  DialogActions,
  DialogContent,
  DialogContentText,
  DialogTitle,
  IconButton,
  Typography,
} from "@material-ui/core";
import DeleteIcon from "@material-ui/icons/Delete";
import React, { FC } from "react";
import Moment from "react-moment";

import { QUERY_USER_SECRETS, REVOKE_USER_SECRET } from "apollo/queries/secret";
import { RevokeUserSecret, RevokeUserSecretVariables } from "apollo/types/RevokeUserSecret";
import { SecretsForUser, SecretsForUserVariables } from "apollo/types/SecretsForUser";
import ContentContainer from "components/ContentContainer";
import { UITable, UITableBody, UITableCell, UITableHead, UITableRow } from "components/UITables";

export interface ListSecretsProps {
  userID: string;
}

const ListSecrets: FC<ListSecretsProps> = ({ userID }) => {
  const [deleteSecretID, setDeleteSecretID] = React.useState<string | undefined>(undefined);

  const { loading, error, data } = useQuery<SecretsForUser, SecretsForUserVariables>(QUERY_USER_SECRETS, {
    variables: { userID },
  });

  const [revokeSecret, { loading: mutLoading }] =
    useMutation<RevokeUserSecret, RevokeUserSecretVariables>(REVOKE_USER_SECRET);

  const dialogue = (
    <Dialog open={!!deleteSecretID}>
      <DialogTitle>Are you sure you want to delete this session?</DialogTitle>
      <DialogContent>
        <DialogContentText>You'll be logged out if you delete your current session.</DialogContentText>
      </DialogContent>
      <DialogActions>
        <Button color="primary" autoFocus onClick={() => setDeleteSecretID(undefined)}>
          No, go back
        </Button>
        <Button
          color="primary"
          autoFocus
          onClick={() => {
            if (deleteSecretID) {
              const secretID = deleteSecretID;
              revokeSecret({
                variables: { secretID },
                update: (cache, { data }) => {
                  if (data && data.revokeUserSecret) {
                    const queryData = cache.readQuery({
                      query: QUERY_USER_SECRETS,
                      variables: { userID },
                    }) as any;
                    const filtered = queryData.secretsForUser.filter((secret: any) => secret.userSecretID !== secretID);
                    cache.writeQuery({
                      query: QUERY_USER_SECRETS,
                      variables: { userID },
                      data: { secretsForUser: filtered },
                    });
                  }
                },
              });
              setDeleteSecretID(undefined);
            }
          }}
        >
          Yes, I'm sure
        </Button>
      </DialogActions>
    </Dialog>
  );

  return (
    <>
      <Typography variant="h2" gutterBottom>
        Browser sessions
      </Typography>
      <Typography variant="body2" color="textSecondary">
        You can delete any session. If you delete your current session, you'll be logged out and will have to log in
        again.
      </Typography>
      <ContentContainer paper margin="normal" loading={loading} error={error && JSON.stringify(error)}>
        <UITable>
          <UITableHead>
            <UITableRow>
              <UITableCell>Created</UITableCell>
              <UITableCell>Exact time</UITableCell>
              <UITableCell>Delete</UITableCell>
            </UITableRow>
          </UITableHead>
          <UITableBody>
            {data?.secretsForUser
              .filter((secret) => secret.description === "Browser session")
              .map(({ createdOn, description, userSecretID }) => (
                <UITableRow key={userSecretID} hover>
                  <UITableCell>
                    <Moment fromNow date={createdOn} />
                  </UITableCell>
                  <UITableCell>{createdOn.toLocaleUpperCase()}</UITableCell>
                  <UITableCell padding="checkbox" align="right">
                    <IconButton
                      aria-label="Delete"
                      disabled={mutLoading}
                      onClick={() => setDeleteSecretID(userSecretID)}
                    >
                      <DeleteIcon />
                    </IconButton>
                  </UITableCell>
                </UITableRow>
              ))}
          </UITableBody>
        </UITable>
        {dialogue}
      </ContentContainer>
    </>
  );
};

export default ListSecrets;
