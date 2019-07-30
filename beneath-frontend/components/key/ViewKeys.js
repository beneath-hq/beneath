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

import { QUERY_PROJECT_KEYS, QUERY_USER_KEYS, REVOKE_KEY } from "../../apollo/queries/key";

const prettyRoles = {
  "r": "Read-only",
  "rw": "Read/write",
  "m": "Browser login",
};

const ViewKeys = ({ entityName, entityID }) => {
  let query = null;
  let queryKey = null;
  let variables = {};
  if (entityName === "user") {
    query = QUERY_USER_KEYS;
    queryKey = "keysForUser";
    variables = { userID: entityID };
  } else if (entityName === "project") {
    query = QUERY_PROJECT_KEYS;
    queryKey = "keysForProject";
    variables = { projectID: entityID };
  }

  return (
    <List dense={true}>
      <Query query={query} variables={variables}>
        {({ loading, error, data }) => {
          if (loading) return <Loading justify="center" />;
          if (error) return <p>Error: {JSON.stringify(error)}</p>;

          let keys = data[queryKey];
          return keys.map(({ createdOn, description, keyID, prefix, role }) => (
            <ListItem key={keyID} disableGutters>
              <ListItemText
                primary={
                  <React.Fragment>
                    {description || <emph>No description</emph>}
                    {" "}(starts with <strong>{prefix}</strong>)
                  </React.Fragment>
                }
                secondary={
                  <React.Fragment>
                    {prettyRoles[role]} key issued <Moment fromNow date={createdOn} />
                  </React.Fragment>
                }
              />
              <ListItemSecondaryAction>
                <Mutation mutation={REVOKE_KEY} update={(cache, { data: { revokeKey } }) => {
                  if (revokeKey) {
                    const queryData = cache.readQuery({ query: query, variables: variables });
                    cache.writeQuery({
                      query: query,
                      variables: variables,
                      data: { [queryKey]: queryData[queryKey].filter((key) => key.keyID !== keyID) },
                    });
                  }
                }}>
                  {(revokeKey, { loading, error }) => (
                    <IconButton edge="end" aria-label="Delete" onClick={() => {
                      revokeKey({ variables: { keyID } });
                    }}>
                      {loading ? <Loading size={20} /> : <DeleteIcon />}
                    </IconButton>
                  )}
                </Mutation>
              </ListItemSecondaryAction>
            </ListItem>
          ))
        }}
      </Query>
    </List>
  );
};

export default ViewKeys;
