import React, { Component } from "react";
import { withRouter } from "next/router";
import { Query } from "react-apollo";

import ExploreSidebar from "../components/ExploreSidebar";
import Loading from "../components/Loading";
import Page from "../components/Page";
import PageTitle from "../components/PageTitle";
import ProfileHero from "../components/ProfileHero";
import SubrouteTabs from "../components/SubrouteTabs";

import EditProject from "../components/pages/project/EditProject";
import ManageMembers from "../components/pages/project/ManageMembers";
import ManageKeys from "../components/pages/key/ManageKeys";
import ViewStreams from "../components/pages/project/ViewStreams";

import withMe from "../hocs/withMe";
import { QUERY_PROJECT } from "../queries/project";

const ProjectPage = ({ router, me }) => (
  <Page title="Project" sidebar={<ExploreSidebar />}>
    <Query query={QUERY_PROJECT} variables={{ name: router.query.name }}>
      {({ loading, error, data }) => {
        if (loading) return <Loading justify="center" />;
        if (error) return <p>Error: {JSON.stringify(error)}</p>;

        console.log("PROJ: ", data)
        let project = data.projectByName;
        let isProjectMember = me && project.users.some((user) => user.userID === me.userID);

        let tabs = [
          { value: "streams", label: "Streams", render: () => (<ViewStreams project={project} />) },
          { value: "members", label: "Members", render: () => (<ManageMembers project={project} editable={isProjectMember} />) },
        ];
        if (isProjectMember) {
          tabs.push({ value: "edit", label: "Edit", render: () => (<EditProject project={project} />) });
          tabs.push({ value: "keys", label: "Keys", render: () => (<ManageKeys projectID={project.projectID} />) });
        }

        return (
          <React.Fragment>
            <PageTitle title={project.displayName} />
            <ProfileHero name={project.displayName} site={project.site}
              description={project.description} avatarURL={project.photoURL}
            />
            <SubrouteTabs defaultValue="streams" tabs={tabs} />
          </React.Fragment>
        );
      }}
    </Query>
  </Page>
);

export default withMe(withRouter(ProjectPage));
