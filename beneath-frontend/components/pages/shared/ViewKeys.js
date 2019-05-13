import React from "react";
import { Query, Mutation } from "react-apollo";

import DeleteIcon from "@material-ui/icons/Delete";
import IconButton from "@material-ui/core/IconButton";
import List from "@material-ui/core/List";
import ListItem from "@material-ui/core/ListItem";
import ListItemText from "@material-ui/core/ListItemText";
import ListItemSecondaryAction from "@material-ui/core/ListItemSecondaryAction";
import Moment from "react-moment";

import Loading from "../../Loading";

import { QUERY_KEYS, REVOKE_KEY } from "../../../queries/key";

const prettyRoles = {
  "r": "Read-only",
  "rw": "Read/write",
  "m": "Browser login",
};

const ViewKeys = ({ userId, projectId }) => {
  let variables = { userId, projectId };
  return (
    <List dense={true}>
      <Query query={QUERY_KEYS} variables={variables}>
        {({ loading, error, data }) => {
          if (loading) return <Loading justify="center" />;
          if (error) return <p>Error: {JSON.stringify(error)}</p>;

          let { keys } = data;
          return keys.map(({ createdOn, description, keyId, prefix, role }) => (
            <ListItem key={keyId} disableGutters>
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
                    const { keys } = cache.readQuery({ query: QUERY_KEYS, variables: variables });
                    cache.writeQuery({
                      query: QUERY_KEYS,
                      variables: variables,
                      data: { keys: keys.filter((key) => key.keyId !== keyId) },
                    });
                  }
                }}>
                  {(revokeKey, { loading, error }) => (
                    <IconButton edge="end" aria-label="Delete" onClick={() => {
                      revokeKey({ variables: { keyId } });
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
