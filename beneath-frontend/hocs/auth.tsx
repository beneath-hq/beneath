import React, { FunctionComponent } from "react";
import Error from "../pages/_error";

interface TokenContextInterface {
  token: string | null;
}

const TokenContext = React.createContext<TokenContextInterface>({
  token: null
});

interface TokenProviderProps {
  token: string;
}

export const TokenProvider: FunctionComponent<TokenProviderProps> = (props) => {
  return (
    <TokenContext.Provider value={{token: props.token}}>
      { props.children }
    </TokenContext.Provider>
  )
}

interface TokenConsumerProps {
  children: (value: TokenContextInterface) => React.ReactNode;
}

export const TokenConsumer: FunctionComponent<TokenConsumerProps> = (props) => {
  return <TokenContext.Consumer>{ props.children }</TokenContext.Consumer>;
}

export const AuthRequired: FunctionComponent = (props) => {
  return (
    <TokenConsumer>
      {({ token }) => {
        if (token) {
          return props.children;
        } else {
          return <Error statusCode={401} />;
        }
      }}
    </TokenConsumer>
  );
}
