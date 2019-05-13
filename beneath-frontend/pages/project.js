import gql from "graphql-tag";
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
import ManageKeys from "../components/pages/shared/ManageKeys";

export const QUERY_PROJECT = gql`
  query Project($name: String) {
    project(name: $name) {
      projectId
      name
      displayName
      site
      description
      createdOn
      updatedOn
      users {
        userId
        name
        username
        photoUrl
      }
      canEdit
    }
  }
`;
// TODO: Get rid of canEdit

const ProjectPage = ({ router }) => (
  <Page title="Project" sidebar={<ExploreSidebar />}>
    <Query query={QUERY_PROJECT} variables={{ name: router.query.name }}>
      {({ loading, error, data }) => {
        if (loading) return <Loading justify="center" />;
        if (error) return <p>Error: {JSON.stringify(error)}</p>;

        let { project } = data;
        let tabs = [
          { value: "members", label: "Members", render: () => (<ViewMembers project={project} canEdit={project.canEdit} />) },
        ];
        if (project.canEdit) {
          tabs.push({ value: "edit", label: "Edit", render: () => (<EditProject projectId={project.projectId} />) });
          tabs.push({ value: "keys", label: "Keys", render: () => (<ManageKeys projectId={project.projectId} />) });
        }

        return (
          <React.Fragment>
            <ProfileHero name={project.displayName} site={project.site}
              description={project.description} avatarUrl={null}
            />
            <SubrouteTabs defaultValue="models" tabs={tabs} />
          </React.Fragment>
        );
      }}
    </Query>
  </Page>
);

export default withRouter(ProjectPage);
