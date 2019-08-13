import { ApolloClient } from "apollo-boost";
import React, { FC } from "react";
import { Query, withApollo, WithApolloClient } from "react-apollo";
import { SubscriptionClient } from "subscriptions-transport-ws";

import { createStyles, Theme, withStyles } from "@material-ui/core";
import Button from "@material-ui/core/Button";
import Grid from "@material-ui/core/Grid";
import Typography from "@material-ui/core/Typography";

import { QUERY_LATEST_RECORDS } from "../../apollo/queries/local/records";
import { LatestRecords, LatestRecordsVariables } from "../../apollo/types/LatestRecords";
import { QueryStream } from "../../apollo/types/QueryStream";
import { GATEWAY_URL_WS } from "../../lib/connection";
import RecordsTable from "../RecordsTable";
import VSpace from "../VSpace";
import { Schema } from "./schema";

const FLASH_ROWS_DURATION = 5000;

const styles = (theme: Theme) => createStyles({
  noMoreDataCaption: {
    color: theme.palette.text.disabled,
  },
});

interface StreamLatestProps extends QueryStream {
  setLoading: (loading: boolean) => void;
  classes: {
    noMoreDataCaption: string;
  };
}

interface StreamLatestState {
  error: string | undefined;
  hasMore: boolean;
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
      hasMore: true,
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
        let queryData = self.apollo.cache.readQuery({
          query: QUERY_LATEST_RECORDS,
          variables: apolloVariables,
        }) as any;

        if (!queryData) {
          queryData = {};
        }

        if (!queryData.latestRecords) {
          queryData.latestRecords = {};
        }

        if (!queryData.latestRecords.data) {
          queryData.latestRecords.data = [];
        }

        const recordID = self.schema.makeUniqueIdentifier(result.data);

        queryData.latestRecords.data = queryData.latestRecords.data.filter(
          (record: any) => record.recordID !== recordID
        );

        queryData.latestRecords.data.unshift({
          __typename: "Record",
          recordID,
          data: result.data,
          timestamp: result.data && result.data["@meta"] && result.data["@meta"].timestamp,
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
      limit: 20,
    };

    return (
      <Query<LatestRecords, LatestRecordsVariables>
        query={QUERY_LATEST_RECORDS}
        variables={variables}
        fetchPolicy="cache-and-network"
      >
        {({ loading, error, data, fetchMore }) => {
          const errorMsg =
            error || (data && data.latestRecords && data.latestRecords.error) || this.state.error;
          if (errorMsg) {
            return <p>Error: {JSON.stringify(error)}</p>;
          }

          const records = data ? (data.latestRecords ? data.latestRecords.data : null) : null;

          let moreElem = null;
          if (this.state.hasMore && records && records.length > 0 && records.length % variables.limit === 0) {
            moreElem = (
              <Grid container justify="center">
                <Grid item>
                  <Button
                    variant="outlined"
                    color="primary"
                    disabled={loading}
                    onClick={() => {
                      const before = records[records.length - 1].timestamp;
                      fetchMore({
                        variables: { ...variables, before },
                        updateQuery: (prev, { fetchMoreResult }) => {
                          if (!fetchMoreResult) {
                            return prev;
                          }
                          const prevRecords = prev.latestRecords.data;
                          const newRecords = fetchMoreResult.latestRecords.data;
                          if (!newRecords || newRecords.length === 0) {
                            this.setState({ hasMore: false });
                          }
                          if (!prevRecords || !newRecords) {
                            return prev;
                          }
                          return {
                            latestRecords: {
                              __typename: "RecordsResponse",
                              data: [...prevRecords, ...newRecords],
                              error: null,
                            },
                          };
                        },
                      });
                    }}
                  >
                    Fetch more
                  </Button>
                </Grid>
              </Grid>
            );
          }

          return (
            <>
              <RecordsTable schema={this.schema} records={records} highlightTopN={this.state.flashRows} />
              <VSpace units={4} />
              {this.state.hasMore && moreElem}
              {!moreElem && (
                <Typography className={this.props.classes.noMoreDataCaption} variant="body2" align="center">
                  There's no more latest data available to load
                </Typography>
              )}
            </>
          );
        }}
      </Query>
    );
  }
}

export default withStyles(styles)(withApollo<StreamLatestProps>(StreamLatest));
