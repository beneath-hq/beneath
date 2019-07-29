import ViewKeys from "./ViewKeys";
import IssueKey from "./IssueKey";

export const ManageProjectKeys = ({ projectID }) => (
  <React.Fragment>
    <IssueKey entityName="project" entityID={projectID} />
    <ViewKeys entityName="project" entityID={projectID} />
  </React.Fragment>
);

export const ManageUserKeys = ({ userID }) => (
  <React.Fragment>
    <IssueKey entityName="user" entityID={userID} />
    <ViewKeys entityName="user" entityID={userID} />
  </React.Fragment>
);
