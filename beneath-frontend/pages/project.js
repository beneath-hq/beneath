import gql from "graphql-tag";
import React, { Component } from "react";
import { withRouter } from "next/router";
import { Query } from "react-apollo";

import ExploreSidebar from "../components/ExploreSidebar";
import Loading from "../components/Loading";
import Page from "../components/Page";
import ProfileHero from "../components/ProfileHero";

const QUERY_PROJECT = gql`
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
        name
        photoUrl
      }
    }
  }
`;

const ProjectPage = ({ router }) => (
  <Page title="Project" sidebar={<ExploreSidebar />}>
    <Query query={QUERY_PROJECT} variables={{ name: router.query.name }}>
      {({ loading, error, data }) => {
        if (loading) return <Loading justify="center" />;
        if (error) return <p>Error: {JSON.stringify(error)}</p>;
        let { project } = data;
        return (
          <ProfileHero name={project.displayName} site={project.site}
            description={project.description} avatarUrl={null}
          />
        );
      }}
    </Query>
  </Page>
);

export default withRouter(ProjectPage);
