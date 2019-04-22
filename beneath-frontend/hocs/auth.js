import Error from "../pages/_error";
import PropTypes from "prop-types";
import React from "react";

const AuthContext = React.createContext({
  user: null
});

export class AuthProvider extends React.Component {
  static propTypes = {
    user: PropTypes.object,
  };

  constructor(props) {
    super(props);
    this.state = {
      user: props.user
    };
  }

  render() {
    return (
      <AuthContext.Provider value={this.state}>
        {this.props.children}
      </AuthContext.Provider>
    );
  }
}

export class AuthConsumer extends React.Component {
  static propTypes = {
    children: PropTypes.func.isRequired,
  };

  render() {
    return <AuthContext.Consumer>{this.props.children}</AuthContext.Consumer>;
  }
}

export class AuthRequired extends React.Component {
  render() {
    return (
      <AuthConsumer>
        {({ user }) => {
          if (user) {
            return this.props.children;
          } else {
            return <Error statusCode={401} />;
          }
        }}
      </AuthConsumer>
    );
  }
}

export const withUser = (App) => {
  return class Auth extends React.Component {
    static displayName = "withAuth(App)";

    static async getInitialProps(ctx) {
      let token = readTokenFromCookie(ctx.ctx ? ctx.ctx.req : null);
      let user = null;
      if (token) {
        user = { token };
      }

      let appProps = {};
      if (App.getInitialProps) {
        appProps = await App.getInitialProps({ ...ctx, user });
      }
      return { ...appProps, user };
    }

    constructor(props) {
      super(props);
      this.user = this.props.user;
    }

    render() {
      return <App {...this.props} user={this.user} />;
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
