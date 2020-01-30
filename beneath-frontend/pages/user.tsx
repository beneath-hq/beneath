import { useQuery } from "@apollo/react-hooks";
import { useRouter } from "next/router";
import React from "react";

import { QUERY_USER_BY_USERNAME } from "../apollo/queries/user";
import { QUERY_ORGANIZATION } from "../apollo/queries/organization";
import { UserByUsername, UserByUsernameVariables } from "../apollo/types/UserByUsername";
import { OrganizationByName, OrganizationByNameVariables, OrganizationByName_organizationByName } from "../apollo/types/OrganizationByName";
import { withApollo } from "../apollo/withApollo";
import useMe from "../hooks/useMe";

import ErrorPage from "../components/ErrorPage";
import Loading from "../components/Loading";
import Page from "../components/Page";
import PageTitle from "../components/PageTitle";
import ProfileHero from "../components/ProfileHero";
import SubrouteTabs from "../components/SubrouteTabs";
import EditMe from "../components/user/EditMe";
import IssueSecret from "../components/user/IssueSecret";
import Monitoring from "../components/user/Monitoring";
import ViewProjects from "../components/user/ViewProjects";
import ViewSecrets from "../components/user/ViewSecrets";
import ViewServices from "../components/organization/ViewServices";
import ViewBilling from "../components/organization/billing/ViewBilling";

const UserPage = () => {
  const me = useMe();
  const router = useRouter();

  const username = router.query.name;
  if (typeof username !== "string") {
    return <ErrorPage statusCode={404} />;
  }

  const { loading, error, data } = useQuery<UserByUsername, UserByUsernameVariables>(QUERY_USER_BY_USERNAME, {
    fetchPolicy: "cache-and-network",
    variables: { username },
  });

  var organization: OrganizationByName_organizationByName | null = null
  if (me && me.organization.personal) {
    const { loading, error, data } = useQuery<OrganizationByName, OrganizationByNameVariables>(QUERY_ORGANIZATION, {
      fetchPolicy: "cache-and-network",
      variables: { name: me.organization.name },
    });

    if (loading) {
      return (
        <Page subheader>
          <Loading justify="center" />
        </Page>
      );
    }

    if (error || !data) {
      return <ErrorPage apolloError={error} />;
    }

    organization = data.organizationByName;
  }

  if (loading) {
    return (
      <Page subheader>
        <Loading justify="center" />
      </Page>
    );
  }

  if (error || !data) {
    return <ErrorPage apolloError={error} />;
  }

  const user = data.userByUsername;
  if (!user) {
    return <ErrorPage statusCode={404} />;
  }

  const tabs = [{ value: "projects", label: "Projects", render: () => <ViewProjects user={user} /> }];

  if (me && me.userID === user.userID) {
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

    if (me.organization.personal) {
      if (!organization) {
        return <ErrorPage statusCode={404} />;
      }
      const org: OrganizationByName_organizationByName = organization
    
      tabs.push({ value: "services", label: "Services", render: () => <ViewServices organization={org} />}),
      tabs.push({ value: "billing", label: "Billing", render: () => <ViewBilling /> })
    }
  }

  return (
    <Page title="User" subheader>
      <PageTitle title={user.name} />
      <ProfileHero name={user.name} description={user.bio} avatarURL={user.photoURL} />
      <SubrouteTabs defaultValue="projects" tabs={tabs} />
    </Page>
  );
};

export default withApollo(UserPage);
