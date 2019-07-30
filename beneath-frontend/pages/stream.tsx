import { SingletonRouter, withRouter } from "next/router";
import React, { FC } from "react";
import { Query } from "react-apollo";

import ExploreSidebar from "../components/ExploreSidebar";
import Loading from "../components/Loading";
import ModelHero from "../components/ModelHero";
import Page from "../components/Page";
import PageTitle from "../components/PageTitle";
import ExploreStream from "../components/stream/ExploreStream";
import SubrouteTabs from "../components/SubrouteTabs";

import EditStream from "../components/stream/EditStream";

import { QUERY_STREAM } from "../queries/stream";

import { QueryStream, QueryStreamVariables } from "../types/generated/QueryStream"

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

          // TODO!

          const tabs = [
            { value: "explore", label: "Explore", render: () => <ExploreStream stream={stream} /> },
            { value: "streaming", label: "Streaming", render: () => <p>Streaming here</p> },
            { value: "api", label: "API", render: () => <p>API here</p> },
            { value: "bigquery", label: "BigQuery", render: () => <p>BigQuery here</p> },
            { value: "write", label: "Write", render: () => <p>Write here</p> },
            { value: "edit", label: "Edit", render: () => <EditStream stream={stream} /> },
          ];

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
