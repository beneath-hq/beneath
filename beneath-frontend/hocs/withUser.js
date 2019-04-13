import React from "react";
import { getUser } from "../lib/auth";

export default App => {
  return class Auth extends React.Component {
    static displayName = "withAuth(App)";

    static async getInitialProps({ Component, router, ctx }) {
      const user = getUser(ctx ? ctx.req : null);
      let appProps = {};
      if (App.getInitialProps) {
        appProps = await App.getInitialProps({ Component, router, ctx });
      }
      return { ...appProps, user };
    }

    constructor(props) {
      super(props);
      this.user = this.props.user;
      if (!this.user && process.browser) {
        this.user = getUser();
      }
    }

    render() {
      return <App {...this.props} user={this.user} />;
    }
  };
};
