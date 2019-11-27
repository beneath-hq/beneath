import { useQuery } from "@apollo/react-hooks";
import { useRouter } from "next/router";
import React from "react";

import Loading from "../components/Loading";
import Page from "../components/Page";

import { QUERY_ORGANIZATION } from "../apollo/queries/organization";
import { OrganizationByName, OrganizationByNameVariables } from "../apollo/types/OrganizationByName";
import { withApollo } from "../apollo/withApollo";
import ErrorPage from "../components/ErrorPage";
import useMe from "../hooks/useMe";
import { toBackendName } from "../lib/names";
import PageTitle from "../components/PageTitle";
import ProfileHero from "../components/ProfileHero";
import SubrouteTabs from "../components/SubrouteTabs";
import ViewUsers from "../components/organization/ViewUsers";
import ViewServices from "../components/organization/ViewServices";
import BillingTab from "../components/organization/BillingTab";


const OrganizationPage = () => {
  const me = useMe();
  const router = useRouter();

  if (typeof router.query.name !== "string") {
    return <ErrorPage statusCode={404} />;
  }

  const { loading, error, data } = useQuery<OrganizationByName, OrganizationByNameVariables>(QUERY_ORGANIZATION, {
    fetchPolicy: "cache-and-network",
    variables: { name: toBackendName(router.query.name) },
  });

  if (loading) {
    return (
      <Page title="Organization" subheader>
        <Loading justify="center" />
      </Page>
    );
  }

  if (error || !data) {
    return <ErrorPage apolloError={error} />;
  }

  const organization = data.organizationByName;
  if (!organization) {
    return <ErrorPage statusCode={404} />;
  }

  const tabs = [{ value: "users", label: "Users", render: () => <ViewUsers organization={organization} /> },
    { value: "services", label: "Services", render: () => <ViewServices organization={organization} /> },
    { value: "billing", label: "Billing", render: () => <BillingTab organization={organization} /> }];

  return (
    <Page title="Organization" subheader>
      <PageTitle title={organization.name} />
      <ProfileHero name={organization.name} />
      <SubrouteTabs defaultValue="users" tabs={tabs} />
    </Page>
  );
  
};

export default withApollo(OrganizationPage);