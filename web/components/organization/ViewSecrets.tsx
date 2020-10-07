import { Box, Container, Paper, Typography, useTheme } from "@material-ui/core";
import React, { FC } from "react";

import IssueSecret from "./secrets/IssueSecret";
import ListSecrets from "./secrets/ListSecrets";

export interface ViewSecretsProps {
  userID: string;
}

const ViewSecrets: FC<ViewSecretsProps> = ({ userID }) => {
  return (
    <Container maxWidth="md">
      <IssueSecret userID={userID} />
      <Box m={4} />
      <ListSecrets userID={userID} />
    </Container>
  );
};

export default ViewSecrets;
