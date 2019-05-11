import React from "react";
import gql from "graphql-tag";
import { Query } from "react-apollo";
import { makeStyles } from "@material-ui/core/styles";

import Loading from "../components/Loading";
import Page from "../components/Page";
import ProfileHero from "../components/ProfileHero";
import SubrouteTabs from "../components/SubrouteTabs";
import { AuthRequired } from "../hocs/auth";

import EditProfile from "../components/pages/profile/EditProfile";
import ManageKeys from "../components/pages/profile/ManageKeys";

const QUERY_ME = gql`
  query {
    me {
      userId
      email
      username
      name
      bio
      photoUrl
      createdOn
      updatedOn
      keys {
        keyId
        description
        prefix
        role
        createdOn
      }
    }
  }
`;

const useStyles = makeStyles((theme) => ({
  profileContent: {
    padding: theme.spacing(8, 0, 6),
  },
}));

export default () => {
  const classes = useStyles();
  return (
    <AuthRequired>
      <Page title="Profile">
        <div className={classes.profileContent}>
          <Query query={QUERY_ME}>
            {({ loading, error, data }) => {
              if (loading) return <Loading justify="center" />;
              if (error) return <p>Error: {JSON.stringify(error)}</p>;
              let { me } = data;
              return (
                <React.Fragment>
                  <ProfileHero name={me.name} description={me.bio} avatarUrl={me.photoUrl} />
                  <SubrouteTabs defaultValue="edit" tabs={[
                    { value: "edit", label: "Edit", render: () => (<EditProfile me={me} />) },
                    { value: "keys", label: "Keys", render: () => (<ManageKeys me={me} />) },
                  ]} />
                </React.Fragment>
              );
            }}
          </Query>
        </div>
      </Page>
    </AuthRequired>
  );
};
