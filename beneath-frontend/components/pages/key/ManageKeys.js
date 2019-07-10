import ViewKeys from "./ViewKeys";
import IssueKey from "./IssueKey";

const ManageKeys = ({ projectID, userID }) => (
  <React.Fragment>
    <IssueKey projectID={projectID} userID={userID} />
    <ViewKeys projectID={projectID} userID={userID} />
  </React.Fragment>
);

export default ManageKeys;