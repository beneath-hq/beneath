import React, { Component } from "react";
import Router from "next/router";

import App from "../../components/App";
import { parseLoginAttempt } from "../../lib/auth";
import { AuthConsumer } from "../../hocs/auth";

class AuthCallback extends Component {
  componentDidMount() {
    let user = parseLoginAttempt((user, error) => {
      if (user) {
        this.setUser(user);
        Router.push("/");
      } else {
        Router.push("/auth/failed");
      }
    });
  }

  render() {
    return (
      <App>
        <AuthConsumer>
          {({ setUser }) => {
            this.setUser = setUser;
          }}
        </AuthConsumer>
      </App>
    );
  }
}

export default AuthCallback;
