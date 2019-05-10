import { withRouter } from "next/router";

import ExploreSidebar from "../components/ExploreSidebar";
import Page from "../components/Page";

const ProjectPage = ({ router }) => (
  <Page title="Project" sidebar={<ExploreSidebar />}>
    <div className="section">
      <div className="title">
        <h1>Project {router.query.name}</h1>
      </div>
    </div>
  </Page>
);

export default withRouter(ProjectPage);
