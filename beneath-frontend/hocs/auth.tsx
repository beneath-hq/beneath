import React, { FunctionComponent } from "react";
import { Query } from "react-apollo";

import { GET_TOKEN } from "../apollo/queries/local/token";
import { Token } from "../apollo/types/Token";
import Error from "../pages/_error";

interface ITokenConsumerProps {
  children: (token: string | null) => React.ReactNode;
}

export const TokenConsumer: FunctionComponent<ITokenConsumerProps> = (props) => {
  return (
    <Query<Token> query={GET_TOKEN}>
      {({ loading, error, data }) => {
        if (data) {
          const { token } = data;
          return props.children(token);
        } else {
          return props.children(null);
        }
      }}
    </Query>
  );
};

export const AuthRequired: FunctionComponent = (props) => {
  return (
    <TokenConsumer>
      {(token) => {
        if (token) {
          return props.children;
        } else {
          return <Error statusCode={401} />;
        }
      }}
    </TokenConsumer>
  );
}
