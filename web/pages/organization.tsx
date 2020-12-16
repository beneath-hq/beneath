import { useQuery } from "@apollo/client";
import { useRouter } from "next/router";
import React from "react";

import { QUERY_ORGANIZATION } from "../apollo/queries/organization";
import { OrganizationByName, OrganizationByNameVariables } from "../apollo/types/OrganizationByName";
import { withApollo } from "../apollo/withApollo";
import useMe from "../hooks/useMe";
import { toBackendName, toURLName } from "../lib/names";
import ErrorPage from "../components/ErrorPage";
import Loading from "../components/Loading";
import EditOrganization from "../components/organization/EditOrganization";
import ViewBilling from "../ee/components/organization/billing/ViewBilling";
import ViewMembers from "../components/organization/ViewMembers";
import ViewUsage from "../components/organization/ViewUsage";
import ViewProjects from "../components/organization/ViewProjects";
import ViewSecrets from "../components/organization/ViewSecrets";
import ViewSecurity from "../components/organization/ViewSecurity";
import Page from "../components/Page";
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
  if (me && organizationName === "me" && typeof window !== "undefined") {
    const tab = router.query.tab;
    const realName = me.name;
    if (tab) {
      router.replace(`/organization?organization_name=${realName}&tab=${tab}`, `/${realName}/-/${tab}`);
    } else {
      router.replace(`/organization?organization_name=${realName}`, `/${realName}`);
    }
    return;
  }

  const { loading, error, data } = useQuery<OrganizationByName, OrganizationByNameVariables>(QUERY_ORGANIZATION, {
    variables: { name: organizationName },
    fetchPolicy: "cache-and-network",
  });

  if (loading) {
    return (
      <Page title={organizationName}>
        <Loading justify="center" />
      </Page>
    );
  }

  if (error || !data) {
    return <ErrorPage apolloError={error} />;
  }

  const organization = data.organizationByName;

  const tabs = [];

  tabs.push({
    value: "projects",
    label: "Projects",
    render: () => <ViewProjects organization={organization} />,
  });

  // if we got private info back
  if (organization.__typename === "PrivateOrganization") {
    if (!organization.personalUserID) {
      // if multi-user org
      tabs.push({ value: "members", label: "Members", render: () => <ViewMembers organization={organization} /> });
    }

    tabs.push({ value: "monitoring", label: "Monitoring", render: () => <ViewUsage organization={organization} /> });

    if (organization.permissions.admin) {
      tabs.push({ value: "edit", label: "Edit", render: () => <EditOrganization organization={organization} /> });
      if (organization.personalUserID) {
        const userID = organization.personalUserID;
        tabs.push({
          value: "secrets",
          label: "Secrets",
          render: () => <ViewSecrets userID={userID} />,
        });
        tabs.push({
          value: "security",
          label: "Security",
          render: () => <ViewSecurity userID={userID} />,
        });
      }
      tabs.push({
        value: "billing",
        label: "Billing",
        render: () => <ViewBilling organization={organization} />,
      });
    }
  }

  return (
    <Page title={toURLName(organization.name)}>
      <ProfileHero
        name={toURLName(organization.name)}
        displayName={organization.displayName}
        description={organization.description}
        avatarURL={organization.photoURL}
      />
      <SubrouteTabs defaultValue="projects" tabs={tabs} />
    </Page>
  );
};

export default withApollo(OrganizationPage);
