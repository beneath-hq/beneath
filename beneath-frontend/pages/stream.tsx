import { SingletonRouter, withRouter } from "next/router";
import React, { FC } from "react";
import { Query } from "react-apollo";

import ExploreSidebar from "../components/ExploreSidebar";
import Loading from "../components/Loading";
import ModelHero from "../components/ModelHero";
import Page from "../components/Page";
import PageTitle from "../components/PageTitle";
import EditStream from "../components/stream/EditStream";
import ExploreStream from "../components/stream/ExploreStream";
import SubrouteTabs from "../components/SubrouteTabs";
import WriteStream from "../components/stream/WriteStream";

import { QUERY_STREAM } from "../apollo/queries/stream";
import { QueryStream, QueryStreamVariables } from "../apollo/types/QueryStream";

interface IProps {
  router: SingletonRouter;
}

const StreamPage: FC<IProps> = ({ router }) => {
  const variables: QueryStreamVariables = {
    name: router.query.name as string,
    projectName: router.query.project_name as string,
  };
  return (
    <Page title="Stream" sidebar={<ExploreSidebar me={null} />}>
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
          tabs.push({ value: "explore", label: "Explore", render: () => <ExploreStream stream={stream} /> });

          if (!stream.batch) {
            tabs.push({ value: "streaming", label: "Streaming", render: () => <p>Streaming here</p> });
          }

          tabs.push({ value: "api", label: "API", render: () => <p>API here</p> });
          tabs.push({ value: "bigquery", label: "BigQuery", render: () => <p>BigQuery here</p> });

          if (stream.manual) {
            tabs.push({ value: "write", label: "Write", render: () => <WriteStream stream={stream} /> });
          }

          tabs.push({ value: "edit", label: "Edit", render: () => <EditStream stream={stream} /> });

          return (
            <React.Fragment>
              <PageTitle title={`${stream.project.name}/${stream.name}`} />
              <ModelHero name={stream.name} description={stream.description} />
              <SubrouteTabs defaultValue="explore" tabs={tabs} />
            </React.Fragment>
          );
        }}
      </Query>
    </Page>
  );
};

export default withRouter(StreamPage);
