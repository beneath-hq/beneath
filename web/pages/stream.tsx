import { useQuery } from "@apollo/client";
import { useRouter } from "next/router";
import React, { useEffect } from "react";
import { toBackendName, toURLName } from "../lib/names";

import { QUERY_STREAM, QUERY_STREAM_INSTANCES } from "../apollo/queries/stream";
import { QUERY_PROJECT } from "../apollo/queries/project";
import {
  StreamByOrganizationProjectAndName,
  StreamByOrganizationProjectAndNameVariables,
  StreamByOrganizationProjectAndName_streamByOrganizationProjectAndName_primaryStreamInstance } from "../apollo/types/StreamByOrganizationProjectAndName";
import { ProjectByOrganizationAndName, ProjectByOrganizationAndNameVariables } from "../apollo/types/ProjectByOrganizationAndName";
import {
  StreamInstancesByOrganizationProjectAndStreamName,
  StreamInstancesByOrganizationProjectAndStreamNameVariables,
  StreamInstancesByOrganizationProjectAndStreamName_streamInstancesByOrganizationProjectAndStreamName,
} from "apollo/types/StreamInstancesByOrganizationProjectAndStreamName";
import { withApollo } from "../apollo/withApollo";

import ErrorPage from "../components/ErrorPage";
import Loading from "../components/Loading";
import ModelHero from "../components/ModelHero";
import Page from "../components/Page";
import ExploreStream from "../components/stream/ExploreStream";
import StreamAPI from "../components/stream/StreamAPI";
import ViewMetrics from "../components/stream/ViewMetrics";
import SubrouteTabs, { SubrouteTabProps } from "../components/SubrouteTabs";
import { useMonthlyMetrics } from "components/metrics/hooks";
import { EntityKind } from "apollo/types/globalTypes";

const StreamPage = () => {
  const router = useRouter();
  const [instance, setInstance] = React.useState<
    StreamInstancesByOrganizationProjectAndStreamName_streamInstancesByOrganizationProjectAndStreamName
    | StreamByOrganizationProjectAndName_streamByOrganizationProjectAndName_primaryStreamInstance
  >({
    __typename: "StreamInstance",
    streamInstanceID: "",
    version: 0,
    createdOn: "",
    madePrimaryOn: "",
    madeFinalOn: "",
  });

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

  const {
    loading: loadingProject,
    error: errorProject,
    data: dataProject,
  } = useQuery<ProjectByOrganizationAndName, ProjectByOrganizationAndNameVariables>(QUERY_PROJECT, {
    variables: { organizationName, projectName },
  });

  const { loading: loadingInstances, error: errorInstances, data: dataInstances } = useQuery<
    StreamInstancesByOrganizationProjectAndStreamName,
    StreamInstancesByOrganizationProjectAndStreamNameVariables
  >(QUERY_STREAM_INSTANCES, {
    variables: { organizationName, projectName, streamName },
  });

  useEffect(() => {
    if (data && data.streamByOrganizationProjectAndName && data.streamByOrganizationProjectAndName.primaryStreamInstance) {
      setInstance(data.streamByOrganizationProjectAndName.primaryStreamInstance);
    }
  }, [data]);

  if (loading || loadingProject || loadingInstances) {
    return (
      <Page title={title} subheader>
        <Loading justify="center" />
      </Page>
    );
  }

  if (error || errorProject || errorInstances || !data || !dataProject || !dataInstances) {
    return <ErrorPage apolloError={error} />;
  }

  const stream = data.streamByOrganizationProjectAndName;
  const project = dataProject.projectByOrganizationAndName;
  const instances = dataInstances.streamInstancesByOrganizationProjectAndStreamName;

  const tabs = [];
  tabs.push({
    value: "data",
    label: "Data",
    render: (props: SubrouteTabProps) => <ExploreStream stream={stream} instance={instance} permissions={project.public} {...props} />,
  });
  tabs.push({ value: "api", label: "API", render: () => <StreamAPI stream={stream} /> });
  tabs.push({ value: "monitoring", label: "Monitoring", render: () => <ViewMetrics stream={stream} /> });

  const defaultValue = stream.primaryStreamInstanceID ? "data" : "api";

  const handleSetInstance = (instanceID: string) => {
    setInstance(instances.find((instance) => instance.streamInstanceID === instanceID) as
      StreamInstancesByOrganizationProjectAndStreamName_streamInstancesByOrganizationProjectAndStreamName);
    return;
  };

  return (
    <Page title={title} subheader>
      <ModelHero
        name={toURLName(stream.name)}
        project={stream.project.name}
        organization={stream.project.organization.name}
        description={stream.description}
        permissions={project.public} // Q: should I instead add the public tag to the stream.project object?
        currentInstance={instance}
        instances={instances}
        setInstance={handleSetInstance}
        metrics={useMonthlyMetrics(EntityKind.Stream, stream.streamID).total}
      />
      <SubrouteTabs defaultValue={defaultValue} tabs={tabs} />
    </Page>
  );
};

export default withApollo(StreamPage);
