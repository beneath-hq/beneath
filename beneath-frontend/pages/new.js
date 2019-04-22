import { MainSidebar } from "../components/Sidebar";
import Page from "../components/Page";

export default () => (
  <Page title="Explore" sidebar={<MainSidebar />} >
    <div className="section">
      <div className="title">
        <h1>New model</h1>
        {/* 
          1. Show a line of buttons (Model, Chart, Notebook), with only "Model" active and working.
          2. Under "Model", show a numbered list of steps to create a new model.
         */}
      </div>
    </div>
  </Page>
);
