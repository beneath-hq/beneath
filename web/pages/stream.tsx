import { useQuery } from "@apollo/client";
import dynamic from "next/dynamic";
import { useRouter } from "next/router";
import React, { useEffect, useState } from "react";

import { QUERY_STREAM } from "apollo/queries/stream";
import {
  StreamByOrganizationProjectAndName,
  StreamByOrganizationProjectAndNameVariables,
} from "apollo/types/StreamByOrganizationProjectAndName";
import { withApollo } from "apollo/withApollo";
import ErrorPage from "components/ErrorPage";
import Loading from "components/Loading";
import Page from "components/Page";
import StreamAPI from "components/stream/StreamAPI";
import StreamHero from "components/stream/StreamHero";
import SubrouteTabs from "components/SubrouteTabs";
import { StreamInstance } from "components/stream/types";
import ViewUsage from "components/stream/ViewUsage";
import { toBackendName, toURLName } from "lib/names";

const DataTab = dynamic(() => import("../components/stream/DataTab"), { ssr: false });

const StreamPage = () => {
  const router = useRouter();
  const [instance, setInstance] = React.useState<StreamInstance | null>(null);
  const [openDialogID, setOpenDialogID] = useState<null | "create" | "promote" | "delete">(null);
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

  // default to the primary instance, if there is one
  useEffect(() => {
    if (data?.streamByOrganizationProjectAndName.primaryStreamInstance) {
      setInstance(data.streamByOrganizationProjectAndName.primaryStreamInstance);
    }
  }, [data?.streamByOrganizationProjectAndName.primaryStreamInstance?.streamInstanceID]);

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
  tabs.push({
    value: "data",
    label: "Data",
    render: () => <DataTab stream={stream} instance={instance} setOpenDialogID={setOpenDialogID} />,
  });
  tabs.push({ value: "api", label: "API", render: () => <StreamAPI stream={stream} /> });
  tabs.push({
    value: "monitoring",
    label: "Monitoring",
    render: () => <>{instance && <ViewUsage stream={stream} instance={instance} />}</>,
  });

  return (
    <Page title={title}>
      <StreamHero
        stream={stream}
        instance={instance || null}
        setInstance={setInstance}
        openDialogID={openDialogID}
        setOpenDialogID={setOpenDialogID}
      />
      <SubrouteTabs defaultValue={"data"} tabs={tabs} />
    </Page>
  );
};

export default withApollo(StreamPage);
