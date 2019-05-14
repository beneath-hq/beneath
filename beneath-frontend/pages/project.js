import React, { Component } from "react";
import { withRouter } from "next/router";
import { Query } from "react-apollo";

import ExploreSidebar from "../components/ExploreSidebar";
import Loading from "../components/Loading";
import Page from "../components/Page";
import ProfileHero from "../components/ProfileHero";
import SubrouteTabs from "../components/SubrouteTabs";

import EditProject from "../components/pages/project/EditProject";
import ViewMembers from "../components/pages/project/ViewMembers";
import ManageKeys from "../components/pages/key/ManageKeys";

import withMe from "../hocs/withMe";
import { QUERY_PROJECT } from "../queries/project";

const ProjectPage = ({ router, me }) => (
  <Page title="Project" sidebar={<ExploreSidebar />}>
    <Query query={QUERY_PROJECT} variables={{ name: router.query.name }}>
      {({ loading, error, data }) => {
        if (loading) return <Loading justify="center" />;
        if (error) return <p>Error: {JSON.stringify(error)}</p>;

        let { project } = data;
        let isProjectMember = me && project.users.some((user) => user.userId === me.userId);

        let tabs = [
          { value: "members", label: "Members", render: () => (<ViewMembers project={project} editable={isProjectMember} />) },
        ];
        if (isProjectMember) {
          tabs.push({ value: "edit", label: "Edit", render: () => (<EditProject projectId={project.projectId} />) });
          tabs.push({ value: "keys", label: "Keys", render: () => (<ManageKeys projectId={project.projectId} />) });
        }

        return (
          <React.Fragment>
            <ProfileHero name={project.displayName} site={project.site}
              description={project.description} avatarUrl={null}
            />
            <SubrouteTabs defaultValue="members" tabs={tabs} />
          </React.Fragment>
        );
      }}
    </Query>
  </Page>
);

export default withMe(withRouter(ProjectPage));

// TODO:
// Add members
// Edit project
