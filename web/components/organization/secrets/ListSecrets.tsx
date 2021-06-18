import { useMutation, useQuery } from "@apollo/client";
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
  Typography,
} from "@material-ui/core";
import DeleteIcon from "@material-ui/icons/Delete";

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
      <DialogTitle>Are you sure you want to delete this secret?</DialogTitle>
      <DialogContent>
        <DialogContentText>
          Any environments (your CLI, any services, Jupyter notebooks, etc.) that rely on this secret will no longer
          work. For the affected environments, you'll have to issue a new secret and re-authenticate.
        </DialogContentText>
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
      <Typography variant="h2">Personal secrets</Typography>
      <ContentContainer paper margin="normal" loading={loading} error={error && JSON.stringify(error)}>
        <UITable>
          <UITableHead>
            <UITableRow>
              <UITableCell>Description</UITableCell>
              <UITableCell>Access</UITableCell>
              <UITableCell>Prefix</UITableCell>
              <UITableCell>Created</UITableCell>
              <UITableCell>Delete</UITableCell>
            </UITableRow>
          </UITableHead>
          <UITableBody>
            {data?.secretsForUser
              .filter((secret) => secret.description !== "Browser session")
              .map(({ createdOn, description, userSecretID, prefix, readOnly, publicOnly }) => (
                <UITableRow key={userSecretID} hover>
                  <UITableCell>{description || ""}</UITableCell>
                  <UITableCell>{readOnly ? (publicOnly ? "Public read" : "Private read") : "Full access"}</UITableCell>
                  <UITableCell>{prefix}</UITableCell>
                  <UITableCell>
                    <Moment fromNow date={createdOn} />
                  </UITableCell>
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
