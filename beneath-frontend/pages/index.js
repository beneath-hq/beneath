import React, { Component } from "react";
import Router from "next/router";

import Page from "../components/Page";
import { TokenConsumer } from "../hocs/auth";

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
        <TokenConsumer>
          {(token) => {
            this.token = token;
            return <></>;
          }}
        </TokenConsumer>
      </Page>
    );
  }
}

export default Index;
