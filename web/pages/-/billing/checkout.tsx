import { useQuery } from "@apollo/client";
import { NextPage } from "next";
import { useRouter } from "next/router";

import { withApollo } from "apollo/withApollo";
import { OrganizationByName, OrganizationByNameVariables } from "apollo/types/OrganizationByName";
import { QUERY_ORGANIZATION } from "apollo/queries/organization";
import ErrorPage from "components/ErrorPage";
import Loading from "components/Loading";
import Page from "components/Page";
import EditBilling from "ee/components/organization/billing/Checkout";
import { IS_EE } from "lib/connection";
import { toBackendName } from "lib/names";
import useMe from "hooks/useMe";

// This page's route appears as /[organization]/-/billing/checkout

const CheckoutPage: NextPage = () => {
  const router = useRouter();
  const me = useMe();

  if (!IS_EE) {
    return <ErrorPage statusCode={404} />;
  }

  if (!me) return null;

  let organizationName;
  if (typeof router.query.organization_name !== "string") {
    organizationName = me.name;
  } else {
    organizationName = toBackendName(router.query.organization_name);
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

  if (organization.__typename === "PrivateOrganization") {
    return (
      <Page title="Checkout" contentMarginTop="normal" maxWidth="md">
        <EditBilling organization={organization} />
      </Page>
    );
  } else {
    return (
      // If a "PublicOrganization" is returned, you don't have permission to access this page
      <ErrorPage />
    );
  }
};

export default withApollo(CheckoutPage);
