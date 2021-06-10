import { NextPage } from "next";

import { withApollo } from "apollo/withApollo";
import AuthToContinue from "components/AuthToContinue";
import Page from "components/Page";
import CreateProject from "components/project/CreateProject";
import { useToken } from "hooks/useToken";

const CreatePage: NextPage = () => {
  const token = useToken();
  return (
    <Page title="Create project" contentMarginTop="normal" maxWidth="sm">
      {!token && <AuthToContinue label="Log in or create a free user to create a project" />}
      {token && <CreateProject />}
    </Page>
  );
};

export default withApollo(CreatePage);
