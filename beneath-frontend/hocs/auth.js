import Error from "../pages/_error";
import PropTypes from "prop-types";
import React from "react";

const TokenContext = React.createContext({
  token: null
});

export class TokenProvider extends React.Component {
  static propTypes = {
    token: PropTypes.string,
  };

  constructor(props) {
    super(props);
    this.state = {
      token: props.token
    };
  }

  render() {
    return (
      <TokenContext.Provider value={this.state}>
        {this.props.children}
      </TokenContext.Provider>
    );
  }
}

export class TokenConsumer extends React.Component {
  static propTypes = {
    children: PropTypes.func.isRequired,
  };

  render() {
    return <TokenContext.Consumer>{this.props.children}</TokenContext.Consumer>;
  }
}

export class AuthRequired extends React.Component {
  render() {
    return (
      <TokenConsumer>
        {({ token }) => {
          if (token) {
            return this.props.children;
          } else {
            return <Error statusCode={401} />;
          }
        }}
      </TokenConsumer>
    );
  }
}

export const withToken = (App) => {
  return class Auth extends React.Component {
    static displayName = "withToken(App)";

    static async getInitialProps(ctx) {
      let token = readTokenFromCookie(ctx.ctx ? ctx.ctx.req : null);

      let appProps = {};
      if (App.getInitialProps) {
        appProps = await App.getInitialProps({ ...ctx, token });
      }
      return { ...appProps, token };
    }

    constructor(props) {
      super(props);
      this.token = props.token;
    }

    render() {
      return <App {...this.props} token={this.token} />;
    }
  };
};

const readTokenFromCookie = (maybeReq) => {
  let token = null;
  let cookie = null;
  if (maybeReq) {
    cookie = maybeReq.headers.cookie;
  } else if (document) {
    cookie = document.cookie;
  }
  if (cookie) {
    cookie = cookie.split(";").find((c) => c.trim().startsWith("token="));
    if (cookie) {
      token = cookie.split("=")[1];
      if (token.length == 0) {
        token = null;
      }
    }
  }
  return token;
};
