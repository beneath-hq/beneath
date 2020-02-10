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


const OrganizationPage = () => {
  const me = useMe();
  const router = useRouter();

  if (!me) {
    return <ErrorPage message={"You must log in to view your organization"} />
  }

  if (typeof router.query.name !== "string") {
    return <ErrorPage statusCode={404} />;
  }

  const organizationName = toBackendName(router.query.name)

  if (me.organization.name !== organizationName) {
    return <ErrorPage message={"You are not a member of an organization named '" + organizationName + "'."} />
  }

  // const tabs = [{ value: "users", label: "Users", render: () => <ViewUsers organization={organization} /> },
  //   { value: "services", label: "Services", render: () => <ViewServices organization={organization} /> },
  //   { value: "billing", label: "Billing", render: () => <ViewBilling organization={organization} /> }];
  const tabs = [{ value: "billing", label: "Billing", render: () => <ViewBilling organizationID={me.organization.organizationID}/> }];

  return (
    <Page title="Organization" subheader>
      <PageTitle title={me.organization.name} />
      <ProfileHero name={me.organization.name} />
      <SubrouteTabs defaultValue="billing" tabs={tabs} />
    </Page>
  );
};

export default withApollo(OrganizationPage);