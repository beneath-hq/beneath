import { useRouter } from "next/router";
import React from "react";
import { useQuery } from "react-apollo";

import Loading from "../components/Loading";
import Page from "../components/Page";
import PageTitle from "../components/PageTitle";
import ProfileHero from "../components/ProfileHero";
import SubrouteTabs from "../components/SubrouteTabs";

import EditProject from "../components/project/EditProject";
import ViewStreams from "../components/project/ViewStreams";

import { QUERY_PROJECT } from "../apollo/queries/project";
import { ProjectByName, ProjectByNameVariables } from "../apollo/types/ProjectByName";
import useMe from "../hocs/useMe";
import { toBackendName, toURLName } from "../lib/names";
import ErrorPage from "../pages/_error";

const ProjectPage = () => {
  const router = useRouter();
  if (typeof router.query.name !== "string") {
    return <ErrorPage statusCode={404} />;
  }

  const me = useMe();
  const { loading, error, data } = useQuery<ProjectByName, ProjectByNameVariables>(QUERY_PROJECT, {
    fetchPolicy: "cache-and-network",
    variables: { name: toBackendName(router.query.name) },
  });

  if (loading) {
    return <Loading justify="center" />;
  }

  if (error || !data) {
    return <ErrorPage apolloError={error} />;
  }

  const project = data.projectByName;
  if (!project) {
    return <ErrorPage statusCode={404} />;
  }

  const isProjectMember = me && project.users.some((user) => user.userID === me.userID);
  const tabs = [{ value: "streams", label: "Streams", render: () => <ViewStreams project={project} /> }];
  if (isProjectMember) {
    tabs.push({ value: "edit", label: "Edit", render: () => <EditProject project={project} /> });
  }

  return (
    <Page title="Project" subheader>
      <PageTitle title={project.displayName || toURLName(project.name)} />
      <ProfileHero
        name={project.displayName || toURLName(project.name)}
        site={project.site}
        description={project.description}
        avatarURL={project.photoURL}
      />
      <SubrouteTabs defaultValue="streams" tabs={tabs} />
    </Page>
  );
};

export default ProjectPage;