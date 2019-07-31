import React from "react";
import { withRouter } from "next/router";
import { Query } from "react-apollo";
import { makeStyles } from "@material-ui/core/styles";

import ExploreSidebar from "../components/ExploreSidebar";
import Loading from "../components/Loading";
import Page from "../components/Page";
import PageTitle from "../components/PageTitle";
import ProfileHero from "../components/ProfileHero";
import SubrouteTabs from "../components/SubrouteTabs";

import EditMe from "../components/user/EditMe";
import ViewProjects from "../components/user/ViewProjects";
import { ManageUserKeys } from "../components/key/ManageKeys";

import withMe from "../hocs/withMe";
import { QUERY_USER } from "../apollo/queries/user";

const useStyles = makeStyles((theme) => ({
}));

const UserPage = ({ router, me }) => {
  const classes = useStyles();
  let userID = router.query.id;
  if (userID === "me") {
    userID = me.userID;
  }
  return (
    <Page title="User" sidebar={<ExploreSidebar />}>
      <div>
        <Query query={QUERY_USER} variables={{ userID }}>
          {({ loading, error, data }) => {
            if (loading) return <Loading justify="center" />;
            if (error) return <p>Error: {JSON.stringify(error)}</p>;
            
            let { user } = data;
            let isMe = userID === "me" || userID === me.userID;
            let tabs = [
              { value: "projects", label: "Projects", render: () => <ViewProjects user={user} /> },
            ];
            if (isMe) {
              tabs.push({ value: "edit", label: "Edit", render: () => <EditMe /> });
              tabs.push({ value: "keys", label: "Keys", render: () => (<ManageUserKeys userID={user.userID} />)});
            }

            return (
              <React.Fragment>
                <PageTitle title={user.name} />
                <ProfileHero name={user.name} description={user.bio} avatarURL={user.photoURL} />
                <SubrouteTabs defaultValue="projects" tabs={tabs} />
              </React.Fragment>
            );
          }}
        </Query>
      </div>
    </Page>
  );
};

export default withMe(withRouter(UserPage));
