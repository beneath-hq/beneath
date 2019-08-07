import { ApolloClient } from "apollo-boost";
import React, { FC } from "react";
import { Query, withApollo } from "react-apollo";
import { SubscriptionClient } from "subscriptions-transport-ws";

import { QUERY_LATEST_RECORDS } from "../../apollo/queries/local/records";
import { LatestRecords, LatestRecordsVariables } from "../../apollo/types/LatestRecords";
import { QueryStream } from "../../apollo/types/QueryStream";
import { GATEWAY_URL_WS } from "../../lib/connection";
import RecordsTable from "../RecordsTable";
import { Schema } from "./schema";

type StreamLatestProps = QueryStream & { client: ApolloClient<any> };

interface StreamLatestState {
  messages: string[];
}

class StreamLatest extends React.Component<StreamLatestProps, StreamLatestState> {
  private apollo: ApolloClient<any>
  private subscription: SubscriptionClient | undefined;
  private schema: Schema;

  constructor(props: StreamLatestProps) {
    super(props);
    this.apollo = props.client;
    this.schema = new Schema(props.stream, true);
    this.state = {
      messages: [],
    };
  }

  public componentWillUnmount() {
    if (this.subscription) {
      this.subscription.close(true);
    }
  }

  public componentDidMount() {
    const self = this;

    this.subscription = new SubscriptionClient(`${GATEWAY_URL_WS}/ws`, {
      reconnect: true,
    });

    const request = {
      query: this.props.stream.currentStreamInstanceID || undefined,
    };

    const apolloVariables = {
      projectName: this.props.stream.project.name,
      streamName: this.props.stream.name,
    };

    // request updates on instance
    this.subscription.request(request).subscribe({
      next: (result) => {
        const queryData = self.apollo.cache.readQuery({
          query: QUERY_LATEST_RECORDS,
          variables: apolloVariables,
        }) as any;

        if (!queryData.latestRecords) {
          queryData.latestRecords = [];
        }

        const recordID = self.schema.makeUniqueIdentifier(result.data);

        queryData.latestRecords = queryData.latestRecords.filter((record: any) => record.recordID !== recordID);

        queryData.latestRecords.unshift({
          __typename: "Record",
          recordID,
          data: result.data,
          sequenceNumber: result.data && result.data["@meta"] && result.data["@meta"].sequence_number,
        });

        self.apollo.writeQuery({
          query: QUERY_LATEST_RECORDS,
          variables: apolloVariables,
          data: queryData,
        });
      },
      error: (result) => {
        // self.setState({
        //   messages: self.state.messages.concat(["ERROR: " + JSON.stringify(result)]),
        // });
      },
      complete: () => {
        console.error("Unexpected subscription complete for instance");
      },
    });
  }

  public render() {
    const variables = {
      projectName: this.props.stream.project.name,
      streamName: this.props.stream.name,
      limit: 100,
    };
    
    return (
      <Query<LatestRecords, LatestRecordsVariables>
        query={QUERY_LATEST_RECORDS}
        variables={variables}
        fetchPolicy="cache-and-network"
      >
        {({ loading, error, data }) => {
          // const errorMsg = loading ? null : error ? error.message : data ? data.records.error : null;

          return <RecordsTable schema={this.schema} records={data ? data.latestRecords : null} />;
        }}
      </Query>
    );
  }
}

export default withApollo(StreamLatest);
