import { MainSidebar } from "../components/Sidebar";
import Page from "../components/Page";

export default () => (
  <Page title="Project" sidebar={<MainSidebar />} >
    <div className="section">
      <div className="title">
        <h1>Project</h1>
      </div>
    </div>
  </Page>
);
