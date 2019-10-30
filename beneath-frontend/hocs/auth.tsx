import React, { FC } from "react";
import { Query, useQuery } from "react-apollo";

import { GET_TOKEN } from "../apollo/queries/local/token";
import { Token } from "../apollo/types/Token";
import Error from "../pages/_error";

interface ITokenConsumerProps {
  children: (token: string | null) => JSX.Element;
}

export const TokenConsumer: FC<ITokenConsumerProps> = (props) => {
  const { loading, error, data } = useQuery<Token>(GET_TOKEN);
  if (data) {
    const { token } = data;
    return props.children(token);
  }

  return props.children(null);
};

export const useToken = () => {
  const { loading, error, data } = useQuery<Token>(GET_TOKEN);
  if (data) {
    const { token } = data;
    return token;
  }

  return null;
};

interface IAuthRequiredProps {
  children: JSX.Element;
}

export const AuthRequired: FC<IAuthRequiredProps> = (props) => {
  return (
    <TokenConsumer>
      {(token) => {
        if (token) {
          return props.children;
        }
        return <Error statusCode={401} />;
      }}
    </TokenConsumer>
  );
};
