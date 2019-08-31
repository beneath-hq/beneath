import React, { FunctionComponent } from "react";
import { Query } from "react-apollo";

import { QUERY_ME } from "../apollo/queries/user";
import { TokenConsumer } from "./auth";

import { Me } from "../apollo/types/Me";

export type ExcludeMeProps<P> = Pick<P, Exclude<keyof P, keyof Me>>;

export const withMe = <P extends Me>(Component: React.ComponentType<P>): FunctionComponent<ExcludeMeProps<P>> => {
  return (props: ExcludeMeProps<P>) => (
    <TokenConsumer>
      {(token) => {
        if (token) {
          return (
            <Query<Me> query={QUERY_ME}>
              {({ loading, error, data }) => {
                if (error) {
                  console.log("withMe error: ", error);
                } else if (!loading && data) {
                  return <Component {...props as any} me={data.me} />;
                }
                return null;
              }}
            </Query>
          );
        } else {
          return <Component {...props as any} me={null} />;
        }
      }}
    </TokenConsumer>
  );
};

export default withMe;
