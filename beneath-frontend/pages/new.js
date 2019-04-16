import { MainSidebar } from "../components/Sidebar";
import Page from "../components/Page";

export default () => (
  <Page title="Explore" sidebar={<MainSidebar />} >
    <div className="section">
      <div className="title">
        <h1>New model</h1>
      </div>
    </div>
  </Page>
);
