import React, { FC } from "react";
import { Subscription } from "react-apollo";
import { SubscriptionClient } from "subscriptions-transport-ws";

import { QUERY_RECORDS } from "../../apollo/queries/local/records";
import { QueryStream } from "../../apollo/types/QueryStream";
import Loading from "../Loading";
import RecordsTable from "../RecordsTable";
import { Schema } from "./schema";

const GRAPHQL_ENDPOINT = "ws://localhost:8080/ws";

interface StreamLatestState {
  messages: string[];
}

class StreamLatest extends React.Component<QueryStream, StreamLatestState> {
  private schema: Schema;
  private client: SubscriptionClient | undefined;

  constructor(props: QueryStream) {
    super(props);
    this.schema = new Schema(props.stream);
    this.state = {
      messages: [],
    };
  }

  public componentWillUnmount() {
    if (this.client) {
      console.log("closing");
      this.client.close();
    }
  }

  public componentDidMount() {
    this.client = new SubscriptionClient(GRAPHQL_ENDPOINT, {
      reconnect: false, // TODO: true
    });

    const _this = this;

    const obs = this.client.request({
      query: "af6602e8-0c55-4579-972e-206f90c668b7",
    });

    obs.subscribe({
      next: (result) => {
        _this.setState({
          messages: _this.state.messages.concat(["NEXT: " + JSON.stringify(result)]),
        });
      },
      error: (result) => {
        _this.setState({
          messages: _this.state.messages.concat(["ERROR: " + JSON.stringify(result)]),
        });
      },
      complete: () => {
        _this.setState({
          messages: _this.state.messages.concat(["COMPLETE: ---"]),
        });
      },
    });

  }

  public render() {
    return <p>{JSON.stringify(this.state.messages)}</p>;
  }
};

export default StreamLatest;
