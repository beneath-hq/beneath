import { useQuery } from "@apollo/react-hooks";
import { useRouter } from "next/router";
import React, { FC } from "react";

import { QUERY_STREAM } from "../apollo/queries/stream";
import { QueryStream, QueryStreamVariables } from "../apollo/types/QueryStream";
import { withApollo } from "../apollo/withApollo";
import { toBackendName, toURLName } from "../lib/names";

import Loading from "../components/Loading";
import ModelHero from "../components/ModelHero";
import Page from "../components/Page";
import PageTitle from "../components/PageTitle";
import ExploreStream from "../components/stream/ExploreStream";
import StreamAPI from "../components/stream/StreamAPI";
import StreamLatest from "../components/stream/StreamLatest";
import StreamMetrics from "../components/stream/StreamMetrics";
import WriteStream from "../components/stream/WriteStream";
import SubrouteTabs, { SubrouteTabProps } from "../components/SubrouteTabs";
import ErrorPage from "../pages/_error";

const StreamPage = () => {
  const router = useRouter();
  if (typeof router.query.name !== "string" || typeof router.query.project_name !== "string") {
    return <ErrorPage statusCode={404} />;
  }

  const { loading, error, data } = useQuery<QueryStream, QueryStreamVariables>(QUERY_STREAM, {
    variables: {
      name: toBackendName(router.query.name as string),
      projectName: toBackendName(router.query.project_name as string),
    },
  });

  if (loading) {
    return (
      <Page title="Project" subheader>
        <Loading justify="center" />
      </Page>
    );
  }

  if (error || !data) {
    return <ErrorPage apolloError={error} />;
  }

  const { stream } = data;

  const tabs = [];

  if (stream.currentStreamInstanceID) {
    tabs.push({ value: "lookup", label: "Lookup", render: () => <ExploreStream stream={stream} /> });

    if (!stream.batch) {
      tabs.push({
        value: "streaming",
        label: "Streaming",
        render: (props: SubrouteTabProps) => <StreamLatest stream={stream} {...props} />,
      });
    }
  }

  tabs.push({ value: "api", label: "API", render: () => <StreamAPI stream={stream} /> });

  if (stream.manual && !stream.batch) {
    tabs.push({ value: "write", label: "Write", render: () => <WriteStream stream={stream} /> });
  }

  tabs.push({ value: "monitoring", label: "Monitoring", render: () => <StreamMetrics stream={stream} /> });

  const defaultValue = stream.currentStreamInstanceID ? "lookup" : "api";
  return (
    <Page title="Stream" subheader>
      <PageTitle title={`${toURLName(stream.project.name)}/${toURLName(stream.name)}`} />
      <ModelHero name={toURLName(stream.name)} description={stream.description} />
      <SubrouteTabs defaultValue={defaultValue} tabs={tabs} />
    </Page>
  );
};

export default withApollo(StreamPage);
