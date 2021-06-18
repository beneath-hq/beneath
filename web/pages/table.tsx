import { useQuery } from "@apollo/client";
import { withApollo } from "apollo/withApollo";
import dynamic from "next/dynamic";
import { useRouter } from "next/router";
import React from "react";

import AssessmentIcon from "@material-ui/icons/Assessment";
import CodeIcon from "@material-ui/icons/Code";
import ViewListIcon from "@material-ui/icons/ViewList";

import { QUERY_TABLE, QUERY_TABLE_INSTANCE } from "apollo/queries/table";
import {
  TableInstanceByOrganizationProjectTableAndVersion,
  TableInstanceByOrganizationProjectTableAndVersionVariables,
  TableInstanceByOrganizationProjectTableAndVersion_tableInstanceByOrganizationProjectTableAndVersion_table,
} from "apollo/types/TableInstanceByOrganizationProjectTableAndVersion";
import {
  TableByOrganizationProjectAndName,
  TableByOrganizationProjectAndNameVariables,
} from "apollo/types/TableByOrganizationProjectAndName";
import ErrorPage from "components/ErrorPage";
import Loading from "components/Loading";
import Page from "components/Page";
import TableAPI from "components/table/TableAPI";
import TableHero from "components/table/TableHero";
import SubrouteTabs from "components/SubrouteTabs";
import ViewUsage from "components/table/ViewUsage";
import { TableInstance } from "components/table/types";
import { toBackendName, toURLName } from "lib/names";

const DataTab = dynamic(() => import("../components/table/DataTab"), { ssr: false });

const safeParseInt = (val: any) => {
  if (typeof val === "string") {
    const int = parseInt(val);
    if (!isNaN(int)) {
      return int;
    }
  }
  return null;
};

// Note: this page is made more complicated because we choose which GraphQL query to run based
// on whether or not the user provides a version in the URL

const TablePage = () => {
  const router = useRouter();
  if (
    typeof router.query.organization_name !== "string" ||
    typeof router.query.project_name !== "string" ||
    typeof router.query.table_name !== "string"
  ) {
    return <ErrorPage statusCode={404} />;
  }

  const organizationName = toBackendName(router.query.organization_name);
  const projectName = toBackendName(router.query.project_name);
  const tableName = toBackendName(router.query.table_name);
  const version = safeParseInt(router.query.version);
  const title =
    `${toURLName(organizationName)}/${toURLName(projectName)}/table:${toURLName(tableName)}` +
    (version !== null ? `/${version}` : "");

  const {
    loading: loadingInstance,
    error: errorInstance,
    data: dataInstance,
  } = useQuery<
    TableInstanceByOrganizationProjectTableAndVersion,
    TableInstanceByOrganizationProjectTableAndVersionVariables
  >(QUERY_TABLE_INSTANCE, {
    variables: { organizationName, projectName, tableName, version: version as number },
    skip: version === null,
  });

  const {
    loading: loadingTable,
    error: errorTable,
    data: dataTable,
  } = useQuery<TableByOrganizationProjectAndName, TableByOrganizationProjectAndNameVariables>(QUERY_TABLE, {
    variables: { organizationName, projectName, tableName },
    skip: version !== null,
  });

  let table: TableInstanceByOrganizationProjectTableAndVersion_tableInstanceByOrganizationProjectTableAndVersion_table;
  let instance: TableInstance | null;
  if (version !== null) {
    if (loadingInstance) {
      return (
        <Page title={title}>
          <Loading justify="center" />
        </Page>
      );
    }
    if (errorInstance || !dataInstance) {
      return <ErrorPage apolloError={errorInstance} />;
    }
    table = dataInstance.tableInstanceByOrganizationProjectTableAndVersion.table;
    instance = dataInstance.tableInstanceByOrganizationProjectTableAndVersion;
  } else {
    if (loadingTable) {
      return (
        <Page title={title}>
          <Loading justify="center" />
        </Page>
      );
    }
    if (errorTable || !dataTable) {
      return <ErrorPage apolloError={errorTable} />;
    }
    table = dataTable.tableByOrganizationProjectAndName;
    instance = dataTable.tableByOrganizationProjectAndName.primaryTableInstance;
  }

  const tabs = [];
  tabs.push({
    value: "data",
    label: "Data",
    icon: <ViewListIcon />,
    render: () => <DataTab table={table} instance={instance} />,
  });
  if (instance) {
    tabs.push({ value: "api", label: "API", icon: <CodeIcon />, render: () => <TableAPI table={table} /> });
    tabs.push({
      value: "monitoring",
      label: "Monitoring",
      icon: <AssessmentIcon />,
      render: () => <>{instance && <ViewUsage table={table} instance={instance} />}</>,
    });
  }

  return (
    <Page title={title}>
      <TableHero table={table} instance={instance || null} />
      <SubrouteTabs defaultValue={"data"} tabs={tabs} />
    </Page>
  );
};

export default withApollo(TablePage);
