import { NextPage } from "next";

import { withApollo } from "../../../apollo/withApollo";
import CreateService from "components/service/CreateService";
import Page from "../../../components/Page";

const CreatePage: NextPage = () => {
  return (
    <Page title="Create service" contentMarginTop="normal" maxWidth="sm">
      <CreateService />
    </Page>
  );
};

export default withApollo(CreatePage);
