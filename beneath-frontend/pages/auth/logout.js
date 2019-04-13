import React, { Component } from "react";
import Router from "next/router";

import App from "../../components/App";
import { logout } from "../../lib/auth";
import { AuthConsumer } from "../../hocs/auth";

class Logout extends Component {
  componentDidMount() {
    logout();
    this.setUser(null);
    Router.push("/");
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

export default Logout;
