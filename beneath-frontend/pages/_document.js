// NOTE: This file is not for layout stuff – use App.js for that
// You should really only use it to change Head

import Document, { Head, Main, NextScript } from "next/document";
import React from "react";
import flush from "styled-jsx/server";
import { ServerStyleSheets } from "@material-ui/styles";

import theme from "../lib/theme";

export default class MyDocument extends Document {
  static async getInitialProps(ctx) {
    // For MUI + Next.js, see https://github.com/mui-org/material-ui/blob/master/examples/nextjs/pages/_document.js
    // Render app and page and get the context of the page with collected side effects.
    const sheets = new ServerStyleSheets();
    const originalRenderPage = ctx.renderPage;

    ctx.renderPage = () =>
      originalRenderPage({
        enhanceApp: App => props => sheets.collect(<App {...props} />),
      });

    const initialProps = await Document.getInitialProps(ctx);

    return {
      ...initialProps,
      // Styles fragment is rendered after the app and page rendering finish.
      styles: (
        <React.Fragment>
          {sheets.getStyleElement()}
          {flush() || null}
        </React.Fragment>
      ),
    };
  }

  render() {
    return (
      <html>
        <Head>
          {/* This shoud really be the only place you make any changes */}

          {/* Meta tags */}
          <meta name="robots" content="index, follow" />
          <meta name="viewport" content="minimum-scale=1, initial-scale=1, width=device-width, shrink-to-fit=no" />
          <meta name="theme-color" content={theme.palette.primary.main} />

          {/* Description tag */}
          <meta
            name="description"
            content="Blockchain data science, Ethereum data API, cryptocurrency investigations, dapp analytics"
          />

          {/* <link href="https://fonts.googleapis.com/css?family=IBM+Plex+Mono:300,300i,600" rel="stylesheet" /> */}

          {/* Stripe tag */}
          <script src="https://js.stripe.com/v3/"></script>
        </Head>
        <body>
          <Main />
          <NextScript />
        </body>
      </html>
    );
  }
}
