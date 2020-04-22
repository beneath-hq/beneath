import { NextPage } from "next";
import React from "react";

import { withApollo } from "../apollo/withApollo";
import Page from "../components/Page";
import Springboard from "../components/terminal/Springboard";
import Welcome from "../components/terminal/Welcome";
import useMe from "../hooks/useMe";

const Terminal: NextPage = () => {
  const me = useMe();
  const loggedIn = !!me;

  return (
    <Page title="Terminal" contentMarginTop="normal">
      {loggedIn && <Springboard />}
      {!loggedIn && <Welcome />}
    </Page>
  );
};

export default withApollo(Terminal);
