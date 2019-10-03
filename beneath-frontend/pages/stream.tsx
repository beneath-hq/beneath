import { SingletonRouter, withRouter } from "next/router";
import React, { FC } from "react";
import { Query } from "react-apollo";

import Loading from "../components/Loading";
import ModelHero from "../components/ModelHero";
import Page from "../components/Page";
import PageTitle from "../components/PageTitle";
import ExploreStream from "../components/stream/ExploreStream";
import StreamAPI from "../components/stream/StreamAPI";
import StreamLatest from "../components/stream/StreamLatest";
import WriteStream from "../components/stream/WriteStream";
import SubrouteTabs, { SubrouteTabProps } from "../components/SubrouteTabs";

import { QUERY_STREAM } from "../apollo/queries/stream";
import { QueryStream, QueryStreamVariables } from "../apollo/types/QueryStream";
import { toBackendName, toURLName } from "../lib/names";

interface IProps {
  router: SingletonRouter;
}

const StreamPage: FC<IProps> = ({ router }) => {
  const variables: QueryStreamVariables = {
    name: toBackendName(router.query.name as string),
    projectName: toBackendName(router.query.project_name as string),
  };
  return (
    <Page title="Stream" subheader>
      <Query<QueryStream, QueryStreamVariables> query={QUERY_STREAM} variables={variables}>
        {({ loading, error, data }) => {
          if (loading) {
            return <Loading justify="center" />;
          }
          if (error || data === undefined) {
            return <p>Error: {JSON.stringify(error)}</p>;
          }

          const { stream } = data;

          const tabs = [];
          tabs.push({ value: "lookup", label: "Lookup", render: () => <ExploreStream stream={stream} /> });

          if (!stream.batch) {
            tabs.push({
              value: "streaming",
              label: "Streaming",
              render: ((props: SubrouteTabProps) => <StreamLatest stream={stream} {...props} />),
            });
          }

          tabs.push({ value: "api", label: "API", render: () => <StreamAPI stream={stream} /> });

          if (stream.manual) {
            tabs.push({ value: "write", label: "Write", render: () => <WriteStream stream={stream} /> });
          }

          tabs.push({ value: "metrics", label: "Metrics", render: () => <p>Metrics</p> });

          return (
            <React.Fragment>
              <PageTitle title={`${toURLName(stream.project.name)}/${toURLName(stream.name)}`} />
              <ModelHero name={toURLName(stream.name)} description={stream.description} />
              <SubrouteTabs defaultValue="lookup" tabs={tabs} />
            </React.Fragment>
          );
        }}
      </Query>
    </Page>
  );
};

export default withRouter(StreamPage);
