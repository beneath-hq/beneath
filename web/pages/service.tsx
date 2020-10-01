import { useQuery } from "@apollo/client";
import { useRouter } from "next/router";
import numbro from "numbro";
import React from "react";

import Loading from "../components/Loading";
import Page from "../components/Page";
import ProfileHero from "../components/ProfileHero";
import SubrouteTabs from "../components/SubrouteTabs";

import { QUERY_SERVICE } from "../apollo/queries/service";
import {
  ServiceByOrganizationProjectAndName,
  ServiceByOrganizationProjectAndNameVariables,
} from "../apollo/types/ServiceByOrganizationProjectAndName";
import { withApollo } from "../apollo/withApollo";
import ErrorPage from "../components/ErrorPage";
import ViewMetrics from "../components/service/ViewMetrics";
import { toBackendName, toURLName } from "../lib/names";

const bytesFormat: numbro.Format = { base: "decimal", mantissa: 1, output: "byte" };

const ServicePage = () => {
  const router = useRouter();

  if (
    typeof router.query.organization_name !== "string" ||
    typeof router.query.project_name !== "string" ||
    typeof router.query.service_name !== "string"
  ) {
    return <ErrorPage statusCode={404} />;
  }

  const organizationName = toBackendName(router.query.organization_name);
  const projectName = toBackendName(router.query.project_name);
  const serviceName = toBackendName(router.query.service_name);
  const title = `${toURLName(serviceName)} – Services – ${toURLName(organizationName)}/${toURLName(projectName)}`;

  const { loading, error, data } = useQuery<
    ServiceByOrganizationProjectAndName,
    ServiceByOrganizationProjectAndNameVariables
  >(QUERY_SERVICE, {
    fetchPolicy: "cache-and-network",
    variables: { organizationName, projectName, serviceName },
  });

  if (loading) {
    return (
      <Page title={title}>
        <Loading justify="center" />
      </Page>
    );
  }

  if (error || !data) {
    return <ErrorPage apolloError={error} />;
  }

  const service = data.serviceByOrganizationProjectAndName;

  const tabs = [];
  tabs.push({ value: "monitoring", label: "Monitoring", render: () => <ViewMetrics service={service} /> });

  return (
    <Page title={title}>
      <ProfileHero
        name={toURLName(service.name)}
        description={
          service.description
            ? service.description + " "
            : "" +
              `(Read quota: ${service.readQuota ? numbro(service.readQuota).format(bytesFormat) : "not set"}, ` +
              `Write quota ${service.writeQuota ? numbro(service.writeQuota).format(bytesFormat) : "not set"}, ` +
              `Scan quota ${service.scanQuota ? numbro(service.scanQuota).format(bytesFormat) : "not set"})`
        }
      />
      <SubrouteTabs defaultValue="monitoring" tabs={tabs} />
    </Page>
  );
};

export default withApollo(ServicePage);
