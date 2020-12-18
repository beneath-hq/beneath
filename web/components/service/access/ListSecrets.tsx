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
  makeStyles,
  Typography,
} from "@material-ui/core";
import DeleteIcon from "@material-ui/icons/Delete";

import { QUERY_SERVICE_SECRETS, REVOKE_SERVICE_SECRET } from "apollo/queries/secret";
import { RevokeServiceSecret, RevokeServiceSecretVariables } from "apollo/types/RevokeServiceSecret";
import { SecretsForService, SecretsForServiceVariables } from "apollo/types/SecretsForService";
import ContentContainer, { CallToAction } from "components/ContentContainer";
import { Table, TableBody, TableCell, TableHead, TableRow } from "components/Tables";
import IssueSecret from "./IssueSecret";
import { Alert, AlertTitle } from "@material-ui/lab";

const useStyles = makeStyles((theme) => ({
  newSecretCard: {
    marginTop: theme.spacing(3),
  },
}));

export interface ListSecretsProps {
  serviceID: string;
}

const ListSecrets: FC<ListSecretsProps> = ({ serviceID }) => {
  const classes = useStyles();
  const [newSecretString, setNewSecretString] = React.useState("");
  const [issueSecretDialog, setIssueSecretDialog] = React.useState(false);
  const [deleteSecretID, setDeleteSecretID] = React.useState<string | undefined>(undefined);

  const { loading, error, data } = useQuery<SecretsForService, SecretsForServiceVariables>(QUERY_SERVICE_SECRETS, {
    variables: { serviceID },
  });

  const [revokeSecret, { loading: mutLoading }] = useMutation<RevokeServiceSecret, RevokeServiceSecretVariables>(
    REVOKE_SERVICE_SECRET
  );

  let cta: CallToAction | undefined;
  if (!data?.secretsForService.length) {
    cta = {
      message: `This service currently has no outstanding secrets`,
      buttons: [{ label: "Issue secret", onClick: () => setIssueSecretDialog(true)}] 
    };
  }

  const issueDialog = (
    <Dialog open={issueSecretDialog} onBackdropClick={() => setIssueSecretDialog(false)}>
      <DialogContent>
        <IssueSecret 
          serviceID={serviceID} 
          onCompleted={(newSecretString) => {
            setNewSecretString(newSecretString);
            setIssueSecretDialog(false);
          }}
        />
      </DialogContent>
    </Dialog>
  );

  const revokeDialog = (
    <Dialog open={!!deleteSecretID}>
      <DialogTitle>Are you sure you want to delete this secret?</DialogTitle>
      <DialogContent>
        <DialogContentText>
          A service in production that depends on this secret will no longer work. You'll have to issue a new secret and re-authenticate.
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
                  if (data && data.revokeServiceSecret) {
                    const queryData = cache.readQuery({
                      query: QUERY_SERVICE_SECRETS,
                      variables: { serviceID },
                    }) as any;
                    const filtered = queryData.secretsForService.filter(
                      (secret: any) => secret.serviceSecretID !== secretID
                    );
                    cache.writeQuery({
                      query: QUERY_SERVICE_SECRETS,
                      variables: { serviceID },
                      data: { secretsForService: filtered },
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
      <Typography variant="h2" gutterBottom>Service secrets</Typography>
      <Typography variant="body2">Service secrets should be used when deploying to production or embedding secrets in public code. 
      The associated usage gets counted against both the service's individual quotas and the organization's overall quotas.</Typography>
      {newSecretString !== "" && (
        <Alert severity="success" className={classes.newSecretCard}>
          <AlertTitle>Here is your new secret!</AlertTitle>
          <Typography gutterBottom variant="body2">
            {newSecretString}
          </Typography>
          <Typography>
            The secret will only be shown this once – remember to keep it safe!
          </Typography>
        </Alert>
      )}
      <ContentContainer paper margin="normal" loading={loading} error={error && JSON.stringify(error)} callToAction={cta}>
        <Table>
          <TableHead>
            <TableRow>
              <TableCell>Description</TableCell>
              <TableCell>Prefix</TableCell>
              <TableCell>Created</TableCell>
              <TableCell>Delete</TableCell>
            </TableRow>
          </TableHead>
          <TableBody>
            {data?.secretsForService
              .filter((secret) => secret.description !== "Browser session")
              .map(({ createdOn, description, serviceSecretID, prefix }) => (
                <TableRow key={serviceSecretID} hover>
                  <TableCell>{description || ""}</TableCell>
                  <TableCell>{prefix}</TableCell>
                  <TableCell>
                    <Moment fromNow date={createdOn} />
                  </TableCell>
                  <TableCell padding="checkbox" align="right">
                    <IconButton
                      aria-label="Delete"
                      disabled={mutLoading}
                      onClick={() => setDeleteSecretID(serviceSecretID)}
                    >
                      <DeleteIcon />
                    </IconButton>
                  </TableCell>
                </TableRow>
              ))}
          </TableBody>
        </Table>
        {issueDialog}
        {revokeDialog}
      </ContentContainer>
      {!cta && (
        <Button variant="contained" onClick={() => setIssueSecretDialog(true)}>Create secret</Button>
      )}
    </>
  );
};

export default ListSecrets;
