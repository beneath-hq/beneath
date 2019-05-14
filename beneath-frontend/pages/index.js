import React, { Component } from "react";
import Router from "next/router";

import Page from "../components/Page";
import { AuthConsumer } from "../hocs/auth";

class Index extends Component {
  componentDidMount() {
    if (this.token) {
      Router.replace("/explore");
    } else {
      Router.replace("/about");
    }
  }

  render() {
    return (
      <Page>
        <AuthConsumer>
          {({ token }) => {
            this.token = token;
          }}
        </AuthConsumer>
      </Page>
    );
  }
}

export default Index;
