import { NextPage } from "next";
import { useRouter } from "next/router";
import React from "react";

import { withApollo } from "../apollo/withApollo";
import Page from "../components/Page";
import Springboard from "../components/terminal/Springboard";
import Welcome from "../components/terminal/Welcome";
import useMe from "../hooks/useMe";

interface Props {
  writeHead?: any;
  end?: any;
}

const Terminal: NextPage<Props> = ({ writeHead, end }) => {
  const me = useMe();
  const loggedIn = !!me;

  if (me?.personalUser && !me?.personalUser.consentTerms) {
    if (typeof window !== "undefined") {
      // client-side redirect
      const router = useRouter();
      router.push("/-/welcome");
    } else if (writeHead && end) {
      // server-side redirect
      // console.log("")
      writeHead(302, { Location: "/-/welcome" });
      end();
    }
  }

  return (
    <Page title="Terminal" contentMarginTop="normal">
      {loggedIn && <Springboard />}
      {!loggedIn && <Welcome />}
    </Page>
  );
};

Terminal.getInitialProps = (ctx) => {
  // NOTE: This hack to redirect in the component is horrible, horrible! Fix it once withApollo
  // is changed to Typescript and we can access ctx.apolloClient and get me in getInitialProps.

  // inject res to component for server-side redirect
  return { writeHead: ctx.res?.writeHead.bind(ctx.res), end: ctx.res?.end.bind(ctx.res) };
};

export default withApollo(Terminal);
