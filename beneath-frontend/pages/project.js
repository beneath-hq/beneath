import React, { Component } from "react";
import { useRouter } from "next/router";
import { Query } from "react-apollo";

import Loading from "../components/Loading";
import Page from "../components/Page";
import PageTitle from "../components/PageTitle";
import ProfileHero from "../components/ProfileHero";
import SubrouteTabs from "../components/SubrouteTabs";

import EditProject from "../components/project/EditProject";
import ViewStreams from "../components/project/ViewStreams";

import { QUERY_PROJECT } from "../apollo/queries/project";
import withMe from "../hocs/withMe";
import { toBackendName, toURLName } from "../lib/names";
import ErrorPage from "../pages/_error";

const ProjectPage = ({ me }) => {
  const router = useRouter();
  if (typeof router.query.name !== "string") {
    return <ErrorPage statusCode={404} />;
  }

  return (
    <Page title="Project" subheader>
      <Query
        query={QUERY_PROJECT}
        variables={{ name: toBackendName(router.query.name) }}
        fetchPolicy="cache-and-network"
      >
        {({ loading, error, data }) => {
          if (loading) {
            return <Loading justify="center" />;
          }
          if (error) {
            return <p>Error: {JSON.stringify(error)}</p>;
          }

          let project = data.projectByName;
          let isProjectMember = me && project.users.some((user) => user.userID === me.userID);

          let tabs = [{ value: "streams", label: "Streams", render: () => <ViewStreams project={project} /> }];
          if (isProjectMember) {
            tabs.push({ value: "edit", label: "Edit", render: () => <EditProject project={project} /> });
          }

          return (
            <React.Fragment>
              <PageTitle title={project.displayName || toURLName(project.name)} />
              <ProfileHero
                name={project.displayName || toURLName(project.name)}
                site={project.site}
                description={project.description}
                avatarURL={project.photoURL}
              />
              <SubrouteTabs defaultValue="streams" tabs={tabs} />
            </React.Fragment>
          );
        }}
      </Query>
    </Page>
  );
};

export default withMe(ProjectPage);
