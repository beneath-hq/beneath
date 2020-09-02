import { NextPage } from "next";

import { withApollo } from "../../../apollo/withApollo";
import CreateProject from "../../../components/project/CreateProject";
import Page from "../../../components/Page";

const CreatePage: NextPage = () => {
  return (
    <Page title="Create project" contentMarginTop="normal" maxWidth="sm">
      <CreateProject />
    </Page>
  );
};

export default withApollo(CreatePage);
