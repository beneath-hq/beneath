import { useQuery } from "@apollo/react-hooks";
import { useRouter } from "next/router";
import React from "react";

import { QUERY_STREAM } from "../apollo/queries/stream";
import { StreamByOrganizationProjectAndName, StreamByOrganizationProjectAndNameVariables } from "../apollo/types/StreamByOrganizationProjectAndName";
import { withApollo } from "../apollo/withApollo";
import { toBackendName, toURLName } from "../lib/names";

import ErrorPage from "../components/ErrorPage";
import Loading from "../components/Loading";
import ModelHero from "../components/ModelHero";
import Page from "../components/Page";
import ExploreStream from "../components/stream/ExploreStream";
import StreamAPI from "../components/stream/StreamAPI";
import ViewMetrics from "../components/stream/ViewMetrics";
// import WriteStream from "../components/stream/WriteStream";
import SubrouteTabs, { SubrouteTabProps } from "../components/SubrouteTabs";

const StreamPage = () => {
  const router = useRouter();

  if (
    typeof router.query.organization_name !== "string"
    || typeof router.query.project_name !== "string"
    || typeof router.query.stream_name !== "string"
  ) {
    return <ErrorPage statusCode={404} />;
  }

  const organizationName = toBackendName(router.query.organization_name);
  const projectName = toBackendName(router.query.project_name);
  const streamName = toBackendName(router.query.stream_name);
  const title = `${toURLName(organizationName)}/${toURLName(projectName)}/${toURLName(streamName)}`;

  const {
    loading,
    error,
    data,
  } = useQuery<StreamByOrganizationProjectAndName, StreamByOrganizationProjectAndNameVariables>(QUERY_STREAM, {
    variables: { organizationName, projectName, streamName },
  });

  if (loading) {
    return (
      <Page title={title} subheader>
        <Loading justify="center" />
      </Page>
    );
  }

  if (error || !data) {
    return <ErrorPage apolloError={error} />;
  }

  const stream = data.streamByOrganizationProjectAndName;

  const tabs = [];

  if (stream.currentStreamInstanceID) {
    tabs.push({
      value: "explore",
      label: "Explore",
      render: (props: SubrouteTabProps) => <ExploreStream stream={stream} {...props} />,
    });
  }

  tabs.push({ value: "api", label: "API", render: () => <StreamAPI stream={stream} /> });

  if (stream.manual && !stream.batch) {
    // disable for now
    // must update js client to be able to write data (current local resolvers do not work anymore!)
    // tabs.push({ value: "write", label: "Write", render: () => <WriteStream stream={stream} /> });
  }

  tabs.push({ value: "monitoring", label: "Monitoring", render: () => <ViewMetrics stream={stream} /> });

  const defaultValue = stream.currentStreamInstanceID ? "explore" : "api";
  return (
    <Page title={title} subheader>
      <ModelHero name={toURLName(stream.name)} description={stream.description} />
      <SubrouteTabs defaultValue={defaultValue} tabs={tabs} />
    </Page>
  );
};

export default withApollo(StreamPage);
