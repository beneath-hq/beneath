import { NextPage } from "next";

import { withApollo } from "../../../apollo/withApollo";
import CreateStream from "../../../components/stream/CreateStream";
import Page from "../../../components/Page";

const CreatePage: NextPage = () => {
  return (
    <Page title="Create stream" contentMarginTop="normal" maxWidth="sm">
      <CreateStream />
    </Page>
  );
};

export default withApollo(CreatePage);
