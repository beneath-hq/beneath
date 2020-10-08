import { NextPage } from "next";

import { withApollo } from "../../../apollo/withApollo";
import Page from "../../../components/Page";

const CreatePage: NextPage = () => {
  return (
    <Page title="Create service" contentMarginTop="normal" maxWidth="sm">
    </Page>
  );
};

export default withApollo(CreatePage);
