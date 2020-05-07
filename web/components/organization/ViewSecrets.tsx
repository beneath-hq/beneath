import React, { FC } from "react";

import IssueSecret from "./secrets/IssueSecret";
import ListSecrets from "./secrets/ListSecrets";

export interface ViewSecretsProps {
  userID: string;
}

const ViewSecrets: FC<ViewSecretsProps> = ({ userID }) => {
  return (
    <>
      <IssueSecret userID={userID} />
      <ListSecrets userID={userID} />
    </>
  );
};

export default ViewSecrets;
