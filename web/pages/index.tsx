import { NextPage } from "next";
import { useRouter } from "next/router";
import React from "react";

import { withApollo } from "../apollo/withApollo";
import Auth from "components/Auth";
import Page from "components/Page";
import Springboard from "components/console/Springboard";
import useMe from "../hooks/useMe";
import { checkForRedirectAfterAuth } from "lib/authRedirect";

interface Props {
  writeHead?: any;
  end?: any;
}

const Console: NextPage<Props> = ({ writeHead, end }) => {
  const me = useMe();
  const loggedIn = !!me;

  if (me?.personalUser && !me?.personalUser.consentTerms) {
    if (typeof window !== "undefined") {
      // client-side redirect
      const router = useRouter();
      router.push("/-/welcome");
    } else if (writeHead && end) {
      // server-side redirect
      writeHead(307, {
        Location: "/-/welcome",
        "Cache-Control": "no-store, no-cache, must-revalidate",
      });
      end();
    }
  }

  if (me?.personalUser?.consentTerms) {
    if (typeof window !== "undefined") {
      // client-side redirect
      const redirectAfterAuth = checkForRedirectAfterAuth();
      if (redirectAfterAuth) {
        const router = useRouter();
        router.push(redirectAfterAuth);
        return (
          // without this prop, React complains about different renderings between client-side and server-side
          <div suppressHydrationWarning={true} />
        );
      }
    }
  }

  if (loggedIn) {
    return (
      <Page title="Console" maxWidth="lg" contentMarginTop="normal">
        <Springboard />
      </Page>
    );
  } else {
    return (
      <Page title="Welcome to Beneath" maxWidth="md" contentMarginTop="normal">
        <Auth />
      </Page>
    );
  }
};

Console.getInitialProps = (ctx) => {
  // NOTE: This hack to redirect in the component is horrible, horrible! Fix it once withApollo
  // is changed to Typescript and we can access ctx.apolloClient and get me in getInitialProps.

  // inject res to component for server-side redirect
  return { writeHead: ctx.res?.writeHead.bind(ctx.res), end: ctx.res?.end.bind(ctx.res) };
};

export default withApollo(Console);
