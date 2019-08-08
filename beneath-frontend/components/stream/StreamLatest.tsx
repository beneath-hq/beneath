import { ApolloClient } from "apollo-boost";
import React, { FC } from "react";
import { Query, withApollo, WithApolloClient } from "react-apollo";
import { SubscriptionClient } from "subscriptions-transport-ws";

import { QUERY_LATEST_RECORDS } from "../../apollo/queries/local/records";
import { LatestRecords, LatestRecordsVariables } from "../../apollo/types/LatestRecords";
import { QueryStream } from "../../apollo/types/QueryStream";
import { GATEWAY_URL_WS } from "../../lib/connection";
import RecordsTable from "../RecordsTable";
import { Schema } from "./schema";

const FLASH_ROWS_DURATION = 5000;

interface StreamLatestProps extends QueryStream {
  setLoading: (loading: boolean) => void;
}

interface StreamLatestState {
  error: string | undefined;
  flashRows: number;
}

class StreamLatest extends React.Component<WithApolloClient<StreamLatestProps>, StreamLatestState> {
  private apollo: ApolloClient<any>;
  private subscription: SubscriptionClient | undefined;
  private schema: Schema;

  constructor(props: WithApolloClient<StreamLatestProps>) {
    super(props);
    this.apollo = props.client;
    this.schema = new Schema(props.stream, true);
    this.state = {
      error: undefined,
      flashRows: 0,
    };
  }

  public componentWillUnmount() {
    this.props.setLoading(false);

    if (this.subscription) {
      this.subscription.close(true);
    }
  }

  public componentDidMount() {
    this.props.setLoading(true);

    if (this.subscription) {
      return;
    }

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

        self.setState({
          error: self.state.error,
          flashRows: self.state.flashRows + 1,
        });

        setTimeout(() => {
          self.setState({
            error: self.state.error,
            flashRows: self.state.flashRows - 1,
          });
        }, FLASH_ROWS_DURATION);
      },
      error: (error) => {
        self.setState({
          error: error.message,
          flashRows: self.state.flashRows,
        });
      },
      complete: () => {
        if (!self.state.error) {
          self.setState({
            error: "Unexpected completion of subscription",
            flashRows: self.state.flashRows,
          });
        }
        self.subscription = undefined;
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
          const errorMsg = error || this.state.error;
          if (errorMsg) {
            return <p>Error: {JSON.stringify(error)}</p>;
          }

          return (
            <RecordsTable
              schema={this.schema}
              records={data ? data.latestRecords : null}
              highlightTopN={this.state.flashRows}
            />
          );
        }}
      </Query>
    );
  }
}

export default withApollo<StreamLatestProps>(StreamLatest);
