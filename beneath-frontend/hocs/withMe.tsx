import React, { FunctionComponent } from "react";
import { Query } from "react-apollo";

import { TokenConsumer } from "./auth";
import { QUERY_ME } from "../queries/user";

import { Me } from "../types/generated/Me"

const withMe = <P extends object>(Component: React.ComponentType<P & Me>): FunctionComponent<P> => {
  return (props: P) => (
    <TokenConsumer>
      {({ token }) => {
        if (token) {
          return (
            <Query<Me> query={QUERY_ME}>
              {({ loading, error, data }) => {
                if (error) {
                  console.log("withMe error: ", error);
                } else if (!loading && data) {
                  return <Component {...props} me={data.me} />
                } 
                return null;
              }}
            </Query>
          );
        } else {
          return (
            <Component {...props} me={null} />
          )
        }
      }}
    </TokenConsumer>
  );
}

export default withMe;
