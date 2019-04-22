import Page from "../components/Page";
import { MainSidebar } from "../components/Sidebar";

export default () => (
  <Page title="Explore" sidebar={<MainSidebar />} >
    <div className="section">
      <div className="title">
        <h1>Explore</h1>
      </div>
    </div>
  </Page>
);
