import React, { Component } from "react";
import Router from "next/router";

import App from "../components/App";
import { AuthConsumer } from "../hocs/auth";

class Index extends Component {
  componentDidMount() {
    if (this.user) {
      Router.push("/explore");
    } else {
      Router.push("/about");
    }
  }

  render() {
    return (
      <App>
        <AuthConsumer>
          {({ user }) => {
            this.user = user;
          }}
        </AuthConsumer>
      </App>
    );
  }
}

export default Index;
