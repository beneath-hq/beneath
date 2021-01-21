import { useQuery } from "@apollo/client";
import { withApollo } from "apollo/withApollo";
import dynamic from "next/dynamic";
import { useRouter } from "next/router";
import React from "react";

import { QUERY_STREAM, QUERY_STREAM_INSTANCE } from "apollo/queries/stream";
import {
  StreamInstanceByOrganizationProjectStreamAndVersion,
  StreamInstanceByOrganizationProjectStreamAndVersionVariables,
  StreamInstanceByOrganizationProjectStreamAndVersion_streamInstanceByOrganizationProjectStreamAndVersion_stream,
} from "apollo/types/StreamInstanceByOrganizationProjectStreamAndVersion";
import {
  StreamByOrganizationProjectAndName,
  StreamByOrganizationProjectAndNameVariables,
} from "apollo/types/StreamByOrganizationProjectAndName";
import ErrorPage from "components/ErrorPage";
import Loading from "components/Loading";
import Page from "components/Page";
import StreamAPI from "components/stream/StreamAPI";
import StreamHero from "components/stream/StreamHero";
import SubrouteTabs from "components/SubrouteTabs";
import ViewUsage from "components/stream/ViewUsage";
import { StreamInstance } from "components/stream/types";
import { toBackendName, toURLName } from "lib/names";

const DataTab = dynamic(() => import("../components/stream/DataTab"), { ssr: false });

// Note: this page is made more complicated because we choose which GraphQL query to run based
// on whether or not the user provides a version in the URL

const StreamPage = () => {
  const router = useRouter();
  if (
    typeof router.query.organization_name !== "string" ||
    typeof router.query.project_name !== "string" ||
    typeof router.query.stream_name !== "string"
  ) {
    return <ErrorPage statusCode={404} />;
  }

  const organizationName = toBackendName(router.query.organization_name);
  const projectName = toBackendName(router.query.project_name);
  const streamName = toBackendName(router.query.stream_name);
  const version = typeof router.query.version === "string" ? parseInt(router.query.version) : null;
  const title =
    `${toURLName(organizationName)}/${toURLName(projectName)}/stream:${toURLName(streamName)}` +
    (version ? `/${version}` : "");

  const { loading: loadingInstance, error: errorInstance, data: dataInstance } = useQuery<
    StreamInstanceByOrganizationProjectStreamAndVersion,
    StreamInstanceByOrganizationProjectStreamAndVersionVariables
  >(QUERY_STREAM_INSTANCE, {
    variables: { organizationName, projectName, streamName, version: version as number },
    skip: version === null,
  });

  const { loading: loadingStream, error: errorStream, data: dataStream } = useQuery<
    StreamByOrganizationProjectAndName,
    StreamByOrganizationProjectAndNameVariables
  >(QUERY_STREAM, {
    variables: { organizationName, projectName, streamName },
    skip: version !== null,
  });

  let stream: StreamInstanceByOrganizationProjectStreamAndVersion_streamInstanceByOrganizationProjectStreamAndVersion_stream;
  let instance: StreamInstance | null;
  if (version) {
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
    stream = dataInstance.streamInstanceByOrganizationProjectStreamAndVersion.stream;
    instance = dataInstance.streamInstanceByOrganizationProjectStreamAndVersion;
  } else {
    if (loadingStream) {
      return (
        <Page title={title}>
          <Loading justify="center" />
        </Page>
      );
    }
    if (errorStream || !dataStream) {
      return <ErrorPage apolloError={errorStream} />;
    }
    stream = dataStream.streamByOrganizationProjectAndName;
    instance = dataStream.streamByOrganizationProjectAndName.primaryStreamInstance;
  }

  const tabs = [];
  tabs.push({
    value: "data",
    label: "Data",
    render: () => <DataTab stream={stream} instance={instance} />,
  });
  if (instance) {
    tabs.push({ value: "api", label: "API", render: () => <StreamAPI stream={stream} /> });
    tabs.push({
      value: "monitoring",
      label: "Monitoring",
      render: () => <>{instance && <ViewUsage stream={stream} instance={instance} />}</>,
    });
  }

  return (
    <Page title={title}>
      <StreamHero stream={stream} instance={instance || null} />
      <SubrouteTabs defaultValue={"data"} tabs={tabs} />
    </Page>
  );
};

export default withApollo(StreamPage);
