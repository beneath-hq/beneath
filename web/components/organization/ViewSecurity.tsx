import { Container } from "@material-ui/core";
import React, { FC } from "react";

import ListSessions from "components/organization/secrets/ListSessions";

export interface ViewBrowserSessionsProps {
  userID: string;
}

export const ViewBrowserSessions: FC<ViewBrowserSessionsProps> = ({ userID }) => {
  return (
    <Container maxWidth="md">
      <ListSessions userID={userID} />
    </Container>
  );
};

export default ViewBrowserSessions;
