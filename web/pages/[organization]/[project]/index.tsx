import { useQuery } from "@apollo/react-hooks";
import { useRouter } from "next/router";
import React from "react";

import Loading from "../../../components/Loading";
import Page from "../../../components/Page";
import PageTitle from "../../../components/PageTitle";
import ProfileHero from "../../../components/ProfileHero";
import SubrouteTabs from "../../../components/SubrouteTabs";

import EditProject from "../../../components/project/EditProject";
import ViewStreams from "../../../components/project/ViewStreams";

import { QUERY_PROJECT } from "../../../apollo/queries/project";
import { ProjectByOrganizationAndName, ProjectByOrganizationAndNameVariables } from "../../../apollo/types/ProjectByOrganizationAndName";
import { withApollo } from "../../../apollo/withApollo";
import ErrorPage from "../../../components/ErrorPage";
import useMe from "../../../hooks/useMe";
import { toBackendName, toURLName } from "../../../lib/names";

const ProjectPage = () => {
  const router = useRouter();

  if (typeof router.query.organization !== "string" || typeof router.query.project !== "string") {
    return <ErrorPage statusCode={404} />;
  }

  const me = useMe();
  const { loading, error, data } = useQuery<ProjectByOrganizationAndName, ProjectByOrganizationAndNameVariables>(QUERY_PROJECT, {
    fetchPolicy: "cache-and-network",
    variables: { organizationName: toBackendName(router.query.organization), projectName: toBackendName(router.query.project) },
  });

  if (loading) {
    return (
      <Page title="Project" subheader>
        <Loading justify="center" />
      </Page>
    );
  }

  if (error || !data) {
    return <ErrorPage apolloError={error} />;
  }

  const project = data.projectByOrganizationAndName;
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

export default withApollo(ProjectPage);
