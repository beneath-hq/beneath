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
import CreateStream from "components/stream/CreateStream";

const CreatePage: NextPage = () => {
  // Prepopulate query text if &stream=... url param is set
  const router = useRouter();
  const organizationName = router.query.organization;
  const projectName = router.query.project;
  const skip = !(typeof organizationName === "string" && typeof projectName === "string");

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
    <Page title="Create stream" contentMarginTop="normal" maxWidth="md">
      <CreateStream preselectedProject={data?.projectByOrganizationAndName} />
    </Page>
  );
};

export default withApollo(CreatePage);
