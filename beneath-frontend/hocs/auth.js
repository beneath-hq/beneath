import React from "react";
import PropTypes from "prop-types";
import Error from "../pages/_error";

const AuthContext = React.createContext({
  user: null,
  setUser: () => {},
});

export class AuthProvider extends React.Component {
  static propTypes = {
    user: PropTypes.object,
  };

  constructor(props) {
    super(props);

    this.setUser = user => {
      this.setState({ user });
    };

    this.state = {
      user: props.user,
      setUser: this.setUser,
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
