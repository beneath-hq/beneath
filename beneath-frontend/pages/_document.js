// NOTE: This file is not for layout stuff – use App.js for that
// You should really only use it to change Head

import Document, { Head, Main, NextScript } from "next/document";
import PropTypes from "prop-types";
import React from "react";
import flush from "styled-jsx/server";

export default class MyDocument extends Document {
  static async getInitialProps(ctx) {
    // For MUI + Next.js, see https://github.com/mui-org/material-ui/blob/master/examples/nextjs/pages/_document.js
    // Render app and page and get the context of the page with collected side effects.
    let pageContext;
    const page = ctx.renderPage((Component) => {
      const WrappedComponent = props => {
        pageContext = props.pageContext;
        return <Component {...props} />;
      };

      WrappedComponent.propTypes = {
        pageContext: PropTypes.object.isRequired,
      };

      return WrappedComponent;
    });

    let css;
    if (pageContext) {
      css = pageContext.sheetsRegistry.toString();
    }

    const initialProps = await Document.getInitialProps(ctx);

    return {
      ...initialProps,
      ...page,
      pageContext,
      // Styles fragment is rendered after the app and page rendering finish.
      styles: (
        <React.Fragment>
          <style
            id="jss-server-side"
            // eslint-disable-next-line react/no-danger
            dangerouslySetInnerHTML={{ __html: css }}
          />
          {flush() || null}
        </React.Fragment>
      ),
    };
  }

  render() {
    const { pageContext } = this.props;

    return (
      <html>
        <Head>
          {/* This shoud really be the only place you make any changes */}

          {/* Meta tags */}
          <meta name="robots" content="index, follow" />
          <meta name="viewport" content="minimum-scale=1, initial-scale=1, width=device-width, shrink-to-fit=no" />
          <meta name="theme-color" content={pageContext ? pageContext.theme.palette.primary.main : null} />

          {/* Description tag */}
          <meta
            name="description"
            content="Blockchain data science, Ethereum data API, cryptocurrency investigations, dapp analytics"
          />

          {/* <link href="https://fonts.googleapis.com/css?family=IBM+Plex+Mono:300,300i,600" rel="stylesheet" /> */}
        </Head>
        <body>
          <Main />
          <NextScript />
        </body>
      </html>
    );
  }
}
