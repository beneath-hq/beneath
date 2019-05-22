import React, { Component } from "react";
import { withRouter } from "next/router";
import { Query } from "react-apollo";

import ExploreSidebar from "../components/ExploreSidebar";
import Loading from "../components/Loading";
import ModelHero from "../components/ModelHero";
import Page from "../components/Page";
import PageTitle from "../components/PageTitle";
import SubrouteTabs from "../components/SubrouteTabs";

import { QUERY_STREAM } from "../queries/stream";

const StreamPage = ({ router }) => (
  <Page title="Stream" sidebar={<ExploreSidebar />}>
    <Query query={QUERY_STREAM} variables={{ name: router.query.name, projectName: router.query["project_name"] }}>
      {({ loading, error, data }) => {
        if (loading) return <Loading justify="center" />;
        if (error) return <p>Error: {JSON.stringify(error)}</p>;

        let { stream } = data;

        let tabs = [
          { value: "explore", label: "Explore", render: () => (<p>Explore here</p>) },
        ];

        return (
          <React.Fragment>
            <PageTitle title={`${stream.name}`} />
            <ModelHero name={stream.name} description={stream.description} />
            <SubrouteTabs defaultValue="explore" tabs={tabs} />
          </React.Fragment>
        );
      }}
    </Query>
  </Page>
);

export default withRouter(StreamPage);
