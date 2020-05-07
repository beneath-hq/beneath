import React, { FC } from "react";

import { Typography } from "@material-ui/core";

import ListSessions from "./secrets/ListSessions";

interface ViewBrowserSessionsProps {
  userID: string;
}

const ViewBrowserSessions: FC<ViewBrowserSessionsProps> = ({ userID }) => {
  return (
    <>
      <Typography variant="h3">Active browser sessions</Typography>
      <Typography>
        You can delete any old sessions or any sessions that you don't recognize.
        Note that if you delete your current session, you'll have to log-in again.
      </Typography>
      <ListSessions userID={userID} />
    </>
  );
};

export default ViewBrowserSessions;
