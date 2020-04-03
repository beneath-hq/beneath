import { useRouter } from "next/router";
import React from "react";
import { withApollo } from "../apollo/withApollo";

import useMe from "../hooks/useMe";
import { toBackendName } from "../lib/names";
import Page from "../components/Page";
import ErrorPage from "../components/ErrorPage";
import PageTitle from "../components/PageTitle";
import ProfileHero from "../components/ProfileHero";
import SubrouteTabs from "../components/SubrouteTabs";
import ViewUsers from "../components/organization/ViewUsers";
import ViewServices from "../components/organization/ViewServices";
import ViewBilling from "../components/organization/billing/ViewBilling";
import ViewOrganizationProjects from "../components/organization/ViewOrganizationProjects";
import { OrganizationByName, OrganizationByNameVariables } from "../apollo/types/OrganizationByName";
import { QUERY_ORGANIZATION } from "../apollo/queries/organization";
import { useQuery } from "@apollo/react-hooks";

const OrganizationPage = () => {
  const me = useMe();
  const router = useRouter();
  
  if (typeof router.query.organization_name !== "string") {
    return <ErrorPage statusCode={404} />;
  }

  const organizationName = toBackendName(router.query.organization_name)

  const { loading, error, data } = useQuery<OrganizationByName, OrganizationByNameVariables>(QUERY_ORGANIZATION, {
    fetchPolicy: "cache-and-network",
    variables: { name: organizationName },
  });

  if (error || !data || !data.organizationByName) {
    return <p>Error: {JSON.stringify(error)}</p>;
  }

  const organization = data.organizationByName

  const tabs = [
    { value: "projects", label: "Projects", render: () => <ViewOrganizationProjects organization={organization} /> },
  ];
  
  if (me && me.organization.name == organizationName) {
    // { value: "users", label: "Users", render: () => <ViewUsers organization={organization} /> },
    // { value: "services", label: "Services", render: () => <ViewServices organization={organization} /> }
    // TODO: billing tab should only be viewable to admins of the organization
    tabs.push({ value: "billing", label: "Billing", render: () => <ViewBilling organizationID={me.organization.organizationID} /> });
  }

  return (
    <Page title="Organization" subheader>
      <PageTitle title={organizationName} />
      <ProfileHero name={organizationName} />
      <SubrouteTabs defaultValue="projects" tabs={tabs} />
    </Page>
  );
};

export default withApollo(OrganizationPage);