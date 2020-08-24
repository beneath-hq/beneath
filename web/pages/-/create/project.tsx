import { Container } from "@material-ui/core";
import { NextPage } from "next";

import { withApollo } from "../../../apollo/withApollo";
import CreateProject from "../../../components/project/CreateProject";
import Page from "../../../components/Page";

const CreatePage: NextPage = () => {
  return (
    <Page title="Create project" contentMarginTop="normal">
      <Container maxWidth="sm">
        <CreateProject />
      </Container>
    </Page>
  );
};

export default withApollo(CreatePage);
