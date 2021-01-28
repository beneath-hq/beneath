// NOTE: This file is not for layout stuff – use App.js for that
// You should really only use it to change Head

import Document, { DocumentContext, Head, Main, NextScript } from "next/document";
import React from "react";
import { ServerStyleSheets } from "@material-ui/styles";

import theme from "../lib/theme";

export default class MyDocument extends Document {
  static async getInitialProps(ctx: DocumentContext) {
    // For MUI + Next.js, see https://github.com/mui-org/material-ui/blob/master/examples/nextjs/pages/_document.js
    // Render app and page and get the context of the page with collected side effects.
    const sheets = new ServerStyleSheets();
    const originalRenderPage = ctx.renderPage;

    ctx.renderPage = () =>
      originalRenderPage({
        enhanceApp: (App) => (props) => sheets.collect(<App {...props} />),
      });

    const initialProps = await Document.getInitialProps(ctx);

    return {
      ...initialProps,
      // Styles fragment is rendered after the app and page rendering finish.
      styles: [...React.Children.toArray(initialProps.styles), sheets.getStyleElement()],
    };
  }

  render() {
    return (
      <html lang="en">
        <Head>
          {/* This should really be the only place you make any changes */}

          {/* Description tag */}
          <meta
            name="description"
            content="Streams you can replay, subscribe, query, and share. We're building a better developer experience for data apps."
          />

          {/* Blog RSS feed */}
          <link
            rel="alternate"
            href="https://about.beneath.dev/feed.xml"
            type="application/rss+xml"
            title="Blog Feed | Beneath"
          />

          {/* Favicon */}
          <meta name="theme-color" content={theme.palette.background.default} />
          <meta name="msapplication-TileColor" content="#10182e" />
          <meta name="msapplication-TileImage" content="/assets/favicon/ms-icon-144x144.png" />
          <meta name="msapplication-config" content="/assets/favicon/browserconfig.xml" />
          <link rel="manifest" href="/assets/favicon/manifest.json" />
          <link rel="apple-touch-icon" sizes="57x57" href="/assets/favicon/apple-icon-57x57.png" />
          <link rel="apple-touch-icon" sizes="60x60" href="/assets/favicon/apple-icon-60x60.png" />
          <link rel="apple-touch-icon" sizes="72x72" href="/assets/favicon/apple-icon-72x72.png" />
          <link rel="apple-touch-icon" sizes="76x76" href="/assets/favicon/apple-icon-76x76.png" />
          <link rel="apple-touch-icon" sizes="114x114" href="/assets/favicon/apple-icon-114x114.png" />
          <link rel="apple-touch-icon" sizes="120x120" href="/assets/favicon/apple-icon-120x120.png" />
          <link rel="apple-touch-icon" sizes="144x144" href="/assets/favicon/apple-icon-144x144.png" />
          <link rel="apple-touch-icon" sizes="152x152" href="/assets/favicon/apple-icon-152x152.png" />
          <link rel="apple-touch-icon" sizes="180x180" href="/assets/favicon/apple-icon-180x180.png" />
          <link rel="icon" type="image/png" sizes="192x192" href="/assets/favicon/android-icon-192x192.png" />
          <link rel="icon" type="image/png" sizes="32x32" href="/assets/favicon/favicon-32x32.png" />
          <link rel="icon" type="image/png" sizes="96x96" href="/assets/favicon/favicon-96x96.png" />
          <link rel="icon" type="image/png" sizes="16x16" href="/assets/favicon/favicon-16x16.png" />
        </Head>
        <body>
          <Main />
          <NextScript />
        </body>
      </html>
    );
  }
}
