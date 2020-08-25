import { Container } from "@material-ui/core";
import { NextPage } from "next";

import { withApollo } from "../../../apollo/withApollo";
import CreateStream from "../../../components/stream/CreateStream";
import Page from "../../../components/Page";

const CreatePage: NextPage = () => {
  return (
    <Page title="Create stream" contentMarginTop="normal">
      <Container maxWidth="sm">
        <CreateStream />
      </Container>
    </Page>
  );
};

export default withApollo(CreatePage);
