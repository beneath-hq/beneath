import React, { Component } from "react";
import { withRouter } from "next/router";
import { Query } from "react-apollo";

import ExploreSidebar from "../components/ExploreSidebar";
import Loading from "../components/Loading";
import ModelHero from "../components/ModelHero";
import Page from "../components/Page";
import PageTitle from "../components/PageTitle";
import SubrouteTabs from "../components/SubrouteTabs";

import EditStream from "../components/pages/stream/EditStream";

import { QUERY_STREAM } from "../queries/stream";

const StreamPage = ({ router }) => (
  <Page title="Stream" sidebar={<ExploreSidebar />}>
    <Query query={QUERY_STREAM} variables={{ name: router.query.name, projectName: router.query["project_name"] }}>
      {({ loading, error, data }) => {
        if (loading) return <Loading justify="center" />;
        if (error) return <p>Error: {JSON.stringify(error)}</p>;

        let { stream } = data;

        // TODO!

        let tabs = [
          { value: "explore", label: "Explore", render: () => (<p>Explore here</p>) },
          { value: "streaming", label: "Streaming", render: () => (<p>Streaming here</p>) },
          { value: "api", label: "API", render: () => (<p>API here</p>) },
          { value: "bigquery", label: "BigQuery", render: () => (<p>BigQuery here</p>) },
          { value: "write", label: "Write", render: () => (<p>Write here</p>) },
          { value: "edit", label: "Edit", render: () => (<EditStream stream={stream} />) },
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

export default withRouter(StreamPage);
