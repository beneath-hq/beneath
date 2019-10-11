import React, { FC } from "react";
import { Mutation, Query, useQuery } from "react-apollo";

import { IconButton, List, ListItem, ListItemSecondaryAction, ListItemText } from "@material-ui/core";
import DeleteIcon from "@material-ui/icons/Delete";
import Moment from "react-moment";

import { QUERY_USER_SECRETS, REVOKE_SECRET } from "../../apollo/queries/secret";
import { RevokeSecret, RevokeSecretVariables } from "../../apollo/types/RevokeSecret";
import { SecretsForUser, SecretsForUserVariables } from "../../apollo/types/SecretsForUser";
import Loading from "../Loading";

interface ViewSecretsProps {
  userID: string;
}

const ViewSecrets: FC<ViewSecretsProps> = ({ userID }) => {
  const { loading, error, data } = useQuery<SecretsForUser, SecretsForUserVariables>(QUERY_USER_SECRETS, {
    variables: { userID },
  });

  if (loading) {
    return <Loading justify="center" />;
  }

  if (error || !data) {
    return <p>Error: {JSON.stringify(error)}</p>;
  }

  return (
    <List dense={true}>
      {data.secretsForUser.map(({ createdOn, description, secretID, prefix }) => (
        <ListItem key={secretID} disableGutters>
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
            <Mutation<RevokeSecret, RevokeSecretVariables>
              mutation={REVOKE_SECRET}
              update={(cache, { data }) => {
                if (data && data.revokeSecret) {
                  const queryData = cache.readQuery({
                    query: QUERY_USER_SECRETS,
                    variables: { userID },
                  }) as any;
                  const filtered = queryData.secretsForUser.filter((secret: any) => secret.secretID !== secretID);
                  cache.writeQuery({
                    query: QUERY_USER_SECRETS,
                    variables: { userID },
                    data: { secretsForUser: filtered },
                  });
                }
              }}
            >
              {(revokeSecret, { loading }) => (
                <IconButton
                  edge="end"
                  aria-label="Delete"
                  onClick={() => {
                    revokeSecret({ variables: { secretID } });
                  }}
                >
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

export default ViewSecrets;
