import React from "react";
import gql from "graphql-tag";
import { withRouter } from "next/router";
import { Query } from "react-apollo";
import { makeStyles } from "@material-ui/core/styles";

import ExploreSidebar from "../components/ExploreSidebar";
import Loading from "../components/Loading";
import Page from "../components/Page";
import ProfileHero from "../components/ProfileHero";
import SubrouteTabs from "../components/SubrouteTabs";

import EditMe from "../components/pages/user/EditMe";
import ManageKeys from "../components/pages/shared/ManageKeys";

const QUERY_USER = gql`
  query User($userId: ID!) {
    user(userId: $userId) {
      userId
      name
      bio
      photoUrl
      createdOn
    }
  }
`;

const useStyles = makeStyles((theme) => ({
}));

const UserPage = ({ router }) => {
  const classes = useStyles();
  const userId = router.query.id;
  return (
    <Page title="User" sidebar={<ExploreSidebar />}>
      <div>
        <Query query={QUERY_USER} variables={{ userId }}>
          {({ loading, error, data }) => {
            if (loading) return <Loading justify="center" />;
            if (error) return <p>Error: {JSON.stringify(error)}</p>;
            
            let { user } = data;
            let tabs = [
              { value: "projects", label: "Projects", render: () => (<p>The projects</p>) },
            ];
            if (userId === "me") {
              tabs.push({ value: "edit", label: "Edit", render: () => <EditMe /> });
              tabs.push({ value: "keys", label: "Keys", render: () => (<ManageKeys userId={user.userId} />)});
            }

            return (
              <React.Fragment>
                <ProfileHero name={user.name} description={user.bio} avatarUrl={user.photoUrl} />
                <SubrouteTabs defaultValue="projects" tabs={tabs} />
              </React.Fragment>
            );
          }}
        </Query>
      </div>
    </Page>
  );
};

export default withRouter(UserPage);
