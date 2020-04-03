import { useQuery } from "@apollo/react-hooks";
import { useRouter } from "next/router";
import React, { FC } from "react";

import { QUERY_STREAM } from "../apollo/queries/stream";
import { StreamByOrganizationProjectAndName, StreamByOrganizationProjectAndNameVariables } from "../apollo/types/StreamByOrganizationProjectAndName";
import { withApollo } from "../apollo/withApollo";
import { toBackendName, toURLName } from "../lib/names";

import ErrorPage from "../components/ErrorPage";
import Loading from "../components/Loading";
import ModelHero from "../components/ModelHero";
import Page from "../components/Page";
import PageTitle from "../components/PageTitle";
import ExploreStream from "../components/stream/ExploreStream";
import StreamAPI from "../components/stream/StreamAPI";
import StreamMetrics from "../components/stream/StreamMetrics";
import WriteStream from "../components/stream/WriteStream";
import SubrouteTabs, { SubrouteTabProps } from "../components/SubrouteTabs";

const StreamPage = () => {
  const router = useRouter();

  if (typeof router.query.organization_name !== "string" || typeof router.query.project_name !== "string" || typeof router.query.stream_name !== "string") {
    return <ErrorPage statusCode={404} />;
  }

  const { loading, error, data } = useQuery<StreamByOrganizationProjectAndName, StreamByOrganizationProjectAndNameVariables>(QUERY_STREAM, {
    variables: {
      organizationName: toBackendName(router.query.organization_name as string),
      projectName: toBackendName(router.query.project_name as string),
      streamName: toBackendName(router.query.stream_name as string),
    },
  });

  if (loading) {
    return (
      <Page title="Stream" subheader>
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
    tabs.push({ value: "write", label: "Write", render: () => <WriteStream stream={stream} /> });
  }

  tabs.push({ value: "monitoring", label: "Monitoring", render: () => <StreamMetrics stream={stream} /> });

  const defaultValue = stream.currentStreamInstanceID ? "explore" : "api";
  return (
    <Page title="Stream" subheader>
      <PageTitle title={`${toURLName(stream.project.name)}/${toURLName(stream.name)}`} />
      <ModelHero name={toURLName(stream.name)} description={stream.description} />
      <SubrouteTabs defaultValue={defaultValue} tabs={tabs} />
    </Page>
  );
};

export default withApollo(StreamPage);
