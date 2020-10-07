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
  Table,
  TableBody,
  TableCell,
  TableHead,
  TableRow,
  Typography,
} from "@material-ui/core";
import DeleteIcon from "@material-ui/icons/Delete";

import { QUERY_USER_SECRETS, REVOKE_USER_SECRET } from "../../../apollo/queries/secret";
import { RevokeUserSecret, RevokeUserSecretVariables } from "../../../apollo/types/RevokeUserSecret";
import { SecretsForUser, SecretsForUserVariables } from "../../../apollo/types/SecretsForUser";
import ContentContainer from "components/ContentContainer";

export interface ListSecretsProps {
  userID: string;
}

const ListSecrets: FC<ListSecretsProps> = ({ userID }) => {
  const [deleteSecretID, setDeleteSecretID] = React.useState<string | undefined>(undefined);

  const { loading, error, data } = useQuery<SecretsForUser, SecretsForUserVariables>(QUERY_USER_SECRETS, {
    variables: { userID },
  });

  const [revokeSecret, { loading: mutLoading }] = useMutation<RevokeUserSecret, RevokeUserSecretVariables>(
    REVOKE_USER_SECRET
  );

  const dialogue = (
    <Dialog open={!!deleteSecretID}>
      <DialogTitle>Are you sure you want to delete this secret?</DialogTitle>
      <DialogContent>
        <DialogContentText>
          Any environments (your CLI, any services, Jupyter notebooks, etc.) that rely on this secret will
          no longer work. For the affected environments, you'll have to issue a new secret and
          re-authenticate.
        </DialogContentText>
      </DialogContent>
      <DialogActions>
        <Button
          color="primary"
          autoFocus
          onClick={() => setDeleteSecretID(undefined)}
        >
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
                    const filtered = queryData.secretsForUser.filter(
                      (secret: any) => secret.userSecretID !== secretID
                    );
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
      <Typography variant="h2" gutterBottom>Personal secrets</Typography>
      <ContentContainer paper loading={loading} error={error && JSON.stringify(error)}>
        <Table>
          <TableHead>
            <TableRow>
              <TableCell>Description</TableCell>
              <TableCell>Access</TableCell>
              <TableCell>Prefix</TableCell>
              <TableCell>Created</TableCell>
              <TableCell>Delete</TableCell>
            </TableRow>
          </TableHead>
          <TableBody>
            {data?.secretsForUser
              .filter((secret) => secret.description !== "Browser session")
              .map(({ createdOn, description, userSecretID, prefix, readOnly, publicOnly }) => (
                <TableRow key={userSecretID} hover>
                  <TableCell>{description}</TableCell> {/* component="th" scope="row" */}
                  <TableCell>{readOnly ? (publicOnly ? "Public read" : "Private read") : "Full access"}</TableCell>
                  <TableCell>{prefix}</TableCell>
                  <TableCell>
                    <Moment fromNow date={createdOn} />
                  </TableCell>
                  <TableCell padding="checkbox" align="right">
                    <IconButton
                      aria-label="Delete"
                      disabled={mutLoading}
                      onClick={() => setDeleteSecretID(userSecretID)}
                    >
                      <DeleteIcon />
                    </IconButton>
                  </TableCell>
                </TableRow>
              ))}
          </TableBody>
        </Table>
        {dialogue}
      </ContentContainer>
    </>
  );
};

export default ListSecrets;
