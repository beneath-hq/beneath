import ExploreSidebar from "../components/ExploreSidebar";
import Page from "../components/Page";

export default () => (
  <Page title="Explore" sidebar={<ExploreSidebar />} >
    <div className="section">
      <div className="title">
        <h1>Explore</h1>
      </div>
    </div>
  </Page>
);
