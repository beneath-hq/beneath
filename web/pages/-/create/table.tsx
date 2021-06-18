import { useQuery } from "@apollo/client";
import { NextPage } from "next";
import { useRouter } from "next/router";

import { QUERY_PROJECT } from "apollo/queries/project";
import {
  ProjectByOrganizationAndName,
  ProjectByOrganizationAndNameVariables,
} from "apollo/types/ProjectByOrganizationAndName";
import { withApollo } from "apollo/withApollo";
import Page from "components/Page";
import CreateTable from "components/table/CreateTable";
import AuthToContinue from "components/AuthToContinue";
import { useToken } from "hooks/useToken";

const CreatePage: NextPage = () => {
  const token = useToken();

  // Prepopulate query text if &table=... url param is set
  const router = useRouter();
  const organizationName = router.query.organization;
  const projectName = router.query.project;
  const skip = !(typeof organizationName === "string" && typeof projectName === "string" && token);

  const { loading, error, data } = useQuery<ProjectByOrganizationAndName, ProjectByOrganizationAndNameVariables>(
    QUERY_PROJECT,
    {
      skip,
      variables: {
        organizationName: typeof organizationName === "string" ? organizationName : "",
        projectName: typeof projectName === "string" ? projectName : "",
      },
    }
  );

  return (
    <Page title="Create table" contentMarginTop="normal" maxWidth="md">
      {!token && <AuthToContinue label="Log in or create a free user to create a table" />}
      {token && <CreateTable preselectedProject={data?.projectByOrganizationAndName} />}
    </Page>
  );
};

export default withApollo(CreatePage);
