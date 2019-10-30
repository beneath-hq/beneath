import React from "react";
import { useRouter } from "next/router";
import { Query } from "react-apollo";
import { makeStyles } from "@material-ui/core/styles";

import Loading from "../components/Loading";
import Page from "../components/Page";
import PageTitle from "../components/PageTitle";
import ProfileHero from "../components/ProfileHero";
import SubrouteTabs from "../components/SubrouteTabs";

import EditMe from "../components/user/EditMe";
import IssueSecret from "../components/user/IssueSecret";
import Monitoring from "../components/user/Monitoring";
import ViewProjects from "../components/user/ViewProjects";
import ViewSecrets from "../components/user/ViewSecrets";
import ErrorPage from "../pages/_error";

import withMe from "../hocs/withMe";
import { QUERY_USER_BY_USERNAME } from "../apollo/queries/user";

const useStyles = makeStyles((theme) => ({}));

const UserPage = ({ me }) => {
  const router = useRouter();
  const username = router.query.name;
  if (typeof username !== "string") {
    return <ErrorPage statusCode={404} />;
  }

  const classes = useStyles();
  return (
    <Page title="User" subheader>
      <div>
        <Query query={QUERY_USER_BY_USERNAME} variables={{ username }} fetchPolicy="cache-and-network">
          {({ loading, error, data }) => {
            if (loading) {
              return <Loading justify="center" />;
            }
            if (error) {
              return <p>Error: {JSON.stringify(error)}</p>;
            }

            let user = data.userByUsername;
            let isMe = user.userID === me.userID;
            let tabs = [{ value: "projects", label: "Projects", render: () => <ViewProjects user={user} /> }];
            if (isMe) {
              tabs.push({ value: "monitoring", label: "Monitoring", render: () => <Monitoring me={me} /> });
              tabs.push({ value: "edit", label: "Edit", render: () => <EditMe /> });
              tabs.push({
                value: "secrets",
                label: "Secrets",
                render: () => (
                  <>
                    <IssueSecret userID={user.userID} />
                    <ViewSecrets userID={user.userID} />
                  </>
                ),
              });
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

export default withMe(UserPage);
