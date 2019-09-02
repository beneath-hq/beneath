import ViewSecrets from "./ViewSecrets";
import IssueSecret from "./IssueSecret";

export const ManageProjectSecrets = ({ projectID }) => (
  <React.Fragment>
    <IssueSecret entityName="project" entityID={projectID} />
    <ViewSecrets entityName="project" entityID={projectID} />
  </React.Fragment>
);

export const ManageUserSecrets = ({ userID }) => (
  <React.Fragment>
    <IssueSecret entityName="user" entityID={userID} />
    <ViewSecrets entityName="user" entityID={userID} />
  </React.Fragment>
);
