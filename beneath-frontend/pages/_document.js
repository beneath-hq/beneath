// NOTE: This file is not for layout stuff – use App.js for that
// You should really only use it to change Head

import Document, { Head, Main, NextScript } from "next/document";
import React from "react";
import flush from "styled-jsx/server";
import { ServerStyleSheets } from '@material-ui/styles';
import * as snippet from '@segment/snippet'

import theme from "../lib/theme";
import { IS_PRODUCTION, SEGMENT_WRITE_KEY } from "../lib/connection";

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

  renderSegmentSnippet() {
    const opts = {
      apiKey: SEGMENT_WRITE_KEY,
      // note: the page option only covers SSR tracking.
      // _app.js is used to track other events using `window.analytics.page()`
      page: true,
    };

    if (IS_PRODUCTION) {
      return snippet.min(opts);
    }
    return snippet.max(opts);
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

          {/* Segment Analytics.js snippet */}
          <script dangerouslySetInnerHTML={{ __html: this.renderSegmentSnippet() }} />
        </Head>
        <body>
          <Main />
          <NextScript />
        </body>
      </html>
    );
  }
}
