import { NextPage } from "next";

import { withApollo } from "../../apollo/withApollo";
import Page from "../../components/Page";
import Main from "../../components/sql/Main";

const SQLPage: NextPage = () => {
  return (
    <Page title="SQL Editor" maxWidth={false}>
      <Main />
    </Page>
  );
};

export default withApollo(SQLPage);
