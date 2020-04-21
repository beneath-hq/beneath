import { useQuery } from "@apollo/react-hooks";
import { useRouter } from "next/router";
import React from "react";

import { QUERY_ORGANIZATION } from "../apollo/queries/organization";
import { QUERY_USER_BY_USERNAME } from "../apollo/queries/user";
import { OrganizationByName, OrganizationByNameVariables } from "../apollo/types/OrganizationByName";
import { UserByUsername, UserByUsernameVariables } from "../apollo/types/UserByUsername";
import { withApollo } from "../apollo/withApollo";
import useMe from "../hooks/useMe";
import { toBackendName } from "../lib/names";

import ErrorPage from "../components/ErrorPage";
import Loading from "../components/Loading";
import BillingTab from "../components/organization/billing/BillingTab";
import EditMe from "../components/organization/personal/EditMe";
import IssueSecret from "../components/organization/personal/IssueSecret";
import Monitoring from "../components/organization/personal/Monitoring";
import ViewBrowserSessions from "../components/organization/personal/ViewBrowserSessions";
import ViewSecrets from "../components/organization/personal/ViewSecrets";
import ViewUserProjects from "../components/organization/personal/ViewUserProjects";
import ViewOrganizationProjects from "../components/organization/ViewOrganizationProjects";
import ViewServices from "../components/organization/ViewServices";
import ViewUsers from "../components/organization/ViewUsers";
import Page from "../components/Page";
import PageTitle from "../components/PageTitle";
import ProfileHero from "../components/ProfileHero";
import SubrouteTabs from "../components/SubrouteTabs";

const OrganizationPage = () => {
  const me = useMe();
  const router = useRouter();

  if (typeof router.query.organization_name !== "string") {
    return <ErrorPage statusCode={404} />;
  }

  const organizationName = toBackendName(router.query.organization_name);

  // little hack to replace organization name "me" with billing organization name if logged in
  if (me) {
    if (organizationName === "me") {
      if (typeof window !== "undefined") {
        const tab = router.query.tab;
        const realName = me?.billingOrganization.name;
        if (tab) {
          router.replace(`/organization?organization_name=${realName}&tab=${tab}`, `/${realName}/-/${tab}`);
        } else {
          router.replace(`/organization?organization_name=${realName}`, `/${realName}`);
        }
        return;
      }
    }
  }

  const { loading, error, data } = useQuery<OrganizationByName, OrganizationByNameVariables>(QUERY_ORGANIZATION, {
    fetchPolicy: "cache-and-network",
    variables: { name: organizationName },
  });

  if (loading) {
    return (
      <Page title="Organization" subheader>
        <Loading justify="center" />
      </Page>
    );
  }

  if (error || !data || !data.organizationByName) {
    return <ErrorPage apolloError={error} />;
  }

  const organization = data.organizationByName;

  const tabs = [
    { value: "projects", label: "Projects", render: () => <ViewOrganizationProjects organization={organization} /> },
  ];

  // logged in
  if (me) {
    if (me.billingOrganization.name === organizationName) {
      // only for your user
      if (organization.personal) {
        const user = organization.users[0];
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
        tabs.push({ value: "security", label: "Security", render: () => <ViewBrowserSessions userID={user.userID} /> });
      } else {
        //  only for your enterprise
        tabs.push({ value: "users", label: "Users", render: () => <ViewUsers organization={organization} /> });
      }
      // { value: "services", label: "Services", render: () => <ViewServices organization={organization} /> }
      // TODO: billing tab should only be viewable to admins of the organization
      tabs.push({
        value: "billing",
        label: "Billing",
        render: () => <BillingTab organizationID={me.billingOrganization.organizationID} />,
      });
    }
  }

  if (organization.personal) {
    const user = organization.users[0];
    return (
      <Page title="User" subheader>
        <PageTitle title={user.name} />
        <ProfileHero name={user.name} description={user.bio} avatarURL={user.photoURL} />
        <SubrouteTabs defaultValue="projects" tabs={tabs} />
      </Page>
    );
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
