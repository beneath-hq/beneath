import { NextPage } from "next";

import { withApollo } from "apollo/withApollo";
import Page from "components/Page";
import Auth from "components/Auth";

const AuthPage: NextPage = () => {
  return (
    <Page title="Welcome to Beneath" contentMarginTop="normal">
      <Auth />
    </Page>
  );
};

export default withApollo(AuthPage);
