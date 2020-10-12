import { useQuery } from "@apollo/client";
import dynamic from "next/dynamic";
import { useRouter } from "next/router";
import React, { useEffect } from "react";

import { QUERY_STREAM } from "../apollo/queries/stream";
import {
  StreamByOrganizationProjectAndName,
  StreamByOrganizationProjectAndNameVariables,
} from "../apollo/types/StreamByOrganizationProjectAndName";
import { withApollo } from "../apollo/withApollo";

import ErrorPage from "../components/ErrorPage";
import Loading from "../components/Loading";
import StreamHero from "../components/stream/StreamHero";
import Page from "../components/Page";
import StreamAPI from "../components/stream/StreamAPI";
import ViewMetrics from "../components/stream/ViewMetrics";
import SubrouteTabs from "../components/SubrouteTabs";
import { toBackendName, toURLName } from "../lib/names";

const ExploreStream = dynamic(() => import("../components/stream/ExploreStream"), { ssr: false });

export interface Instance {
  streamInstanceID: string;
  version: number;
  madePrimaryOn: ControlTime | null;
  madeFinalOn: ControlTime | null;
}

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
  const title = `${toURLName(organizationName)}/${toURLName(projectName)}/${toURLName(streamName)}`;

  const { loading, error, data } = useQuery<
    StreamByOrganizationProjectAndName,
    StreamByOrganizationProjectAndNameVariables
  >(QUERY_STREAM, {
    variables: { organizationName, projectName, streamName },
  });

  const [instance, setInstance] = React.useState<Instance | null>(
    data?.streamByOrganizationProjectAndName.primaryStreamInstance || null
  );

  useEffect(() => {
    if (data?.streamByOrganizationProjectAndName) {
      setInstance(data?.streamByOrganizationProjectAndName.primaryStreamInstance);
    }
  }, [data?.streamByOrganizationProjectAndName.streamID]);

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

  const stream = data.streamByOrganizationProjectAndName;

  const tabs = [];
  tabs.push({ value: "data", label: "Data", render: () => <ExploreStream stream={stream} instance={instance} /> });
  tabs.push({ value: "api", label: "API", render: () => <StreamAPI stream={stream} /> });
  tabs.push({ value: "monitoring", label: "Monitoring", render: () => <ViewMetrics stream={stream} /> });

  return (
    <Page title={title}>
      <StreamHero stream={stream} instance={instance} setInstance={setInstance} />
      <SubrouteTabs defaultValue={"data"} tabs={tabs} />
    </Page>
  );
};

export default withApollo(StreamPage);
