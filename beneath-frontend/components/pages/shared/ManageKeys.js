import ViewKeys from "./ViewKeys";
import IssueKey from "./IssueKey";

const ManageKeys = ({ projectId, userId }) => (
  <React.Fragment>
    <IssueKey projectId={projectId} userId={userId} />
    <ViewKeys projectId={projectId} userId={userId} />
  </React.Fragment>
);

export default ManageKeys;