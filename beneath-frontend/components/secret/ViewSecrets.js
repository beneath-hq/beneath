import React from "react";
import { Query, Mutation } from "react-apollo";

import DeleteIcon from "@material-ui/icons/Delete";
import IconButton from "@material-ui/core/IconButton";
import List from "@material-ui/core/List";
import ListItem from "@material-ui/core/ListItem";
import ListItemText from "@material-ui/core/ListItemText";
import ListItemSecondaryAction from "@material-ui/core/ListItemSecondaryAction";
import Moment from "react-moment";

import Loading from "../Loading";

import {
  QUERY_PROJECT_SECRETS,
  QUERY_USER_SECRETS,
  REVOKE_SECRET
} from "../../apollo/queries/secret";

const prettyRoles = {
  "r": "Read-only",
  "rw": "Read/write",
  "m": "Browser login",
};

const ViewSecrets = ({ entityName, entityID }) => {
  let query = null;
  let queryKey = null;
  let variables = {};
  if (entityName === "user") {
    query = QUERY_USER_SECRETS;
    queryKey = "secretsForUser";
    variables = { userID: entityID };
  } else if (entityName === "project") {
    query = QUERY_PROJECT_SECRETS;
    queryKey = "secretsForProject";
    variables = { projectID: entityID };
  }

  return (
    <List dense={true}>
      <Query query={query} variables={variables}>
        {({ loading, error, data }) => {
          if (loading) return <Loading justify="center" />;
          if (error) return <p>Error: {JSON.stringify(error)}</p>;

          let secrets = data[queryKey];
          return secrets.map(({ createdOn, description, secretID, prefix, role }) => (
            <ListItem key={secretID} disableGutters>
              <ListItemText
                primary={
                  <React.Fragment>
                    {description || <emph>No description</emph>} (starts with <strong>{prefix}</strong>)
                  </React.Fragment>
                }
                secondary={
                  <React.Fragment>
                    {prettyRoles[role]} secret issued <Moment fromNow date={createdOn} />
                  </React.Fragment>
                }
              />
              <ListItemSecondaryAction>
                <Mutation
                  mutation={REVOKE_SECRET}
                  update={(cache, { data: { revokeSecret } }) => {
                    if (revokeSecret) {
                      const queryData = cache.readQuery({ query: query, variables: variables });
                      cache.writeQuery({
                        query: query,
                        variables: variables,
                        data: { [queryKey]: queryData[queryKey].filter((secret) => secret.secretID !== secretID) },
                      });
                    }
                  }}
                >
                  {(revokeSecret, { loading, error }) => (
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
          ));
        }}
      </Query>
    </List>
  );
};

export default ViewSecrets;
