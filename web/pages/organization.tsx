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
import ViewBilling from "../components/organization/billing/ViewBillingInfo";
import ViewUserProjects from "../components/organization/personal/ViewUserProjects";
import ViewOrganizationProjects from "../components/organization/ViewOrganizationProjects";
import { OrganizationByName, OrganizationByNameVariables } from "../apollo/types/OrganizationByName";
import { QUERY_USER_BY_USERNAME } from "../apollo/queries/user";
import { UserByUsername, UserByUsernameVariables } from "../apollo/types/UserByUsername";
import { QUERY_ORGANIZATION } from "../apollo/queries/organization";
import { useQuery } from "@apollo/react-hooks";
import EditMe from "../components/organization/personal/EditMe";
import IssueSecret from "../components/organization/personal/IssueSecret";
import Monitoring from "../components/organization/personal/Monitoring";
import ViewSecrets from "../components/organization/personal/ViewSecrets";
import ViewBrowserSessions from "../components/organization/personal/ViewBrowserSessions";

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

  const tabs = [{ value: "projects", label: "Projects", render: () => <ViewOrganizationProjects organization={organization} /> }]

  // logged in
  if (me) {
    if (me.billingOrganization.name === organizationName) {
      // only for your user
      if (organization.personal) { 
        const user = organization.users[0]
        tabs.push({ value: "monitoring", label: "Monitoring", render: () => <Monitoring me={me} /> })
        tabs.push({ value: "edit", label: "Edit", render: () => <EditMe /> })
        tabs.push({
          value: "secrets",
          label: "Secrets",
          render: () => (
            <>
              <IssueSecret userID={user.userID} />
              <ViewSecrets userID={user.userID} />
            </>
          )
        })
        tabs.push({ value: "security", label: "Security", render: () => <ViewBrowserSessions userID={user.userID} /> })
      } else { //  only for your enterprise
        tabs.push({ value: "users", label: "Users", render: () => <ViewUsers organization={organization} /> })
      }
      // { value: "services", label: "Services", render: () => <ViewServices organization={organization} /> }
      // TODO: billing tab should only be viewable to admins of the organization
      tabs.push({ value: "billing", label: "Billing", render: () => <ViewBilling organizationID={me.billingOrganization.organizationID} /> });
    }
  }

  if (organization.personal) {
    const user = organization.users[0]
    return (
      <Page title="User" subheader>
      <PageTitle title={user.name} />
      <ProfileHero name={user.name} description={user.bio} avatarURL={user.photoURL} />
      <SubrouteTabs defaultValue="projects" tabs={tabs} />
    </Page>
    )
  } else {
    return (
      <Page title="Organization" subheader>
        <PageTitle title={organizationName} />
        <ProfileHero name={organizationName} />
        <SubrouteTabs defaultValue="projects" tabs={tabs} />
      </Page>
    );
  }

};

export default withApollo(OrganizationPage);
