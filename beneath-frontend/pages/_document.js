// NOTE: This file is not for layout stuff – use App.js for that
// You should really only use it to change Head

import Document, { Head, Main, NextScript } from "next/document";

export default class MyDocument extends Document {
  static async getInitialProps(ctx) {
    const initialProps = await Document.getInitialProps(ctx);
    return { ...initialProps };
  }

  render() {
    return (
      <html>
        <Head>
          {/* This shoud really be the only place you make any changes */}

          {/* Global site tag (gtag.js) - Google Analytics */}
          {/* <script async src="https://www.googletagmanager.com/gtag/js?id=UA-118362426-2"></script>
          <script>{`
            window.dataLayer = window.dataLayer || [];
            function gtag(){ dataLayer.push(arguments); }
            gtag('js', new Date());    
            gtag('config', 'UA-118362426-2');
          `}</script> */}

          {/* Meta tags */}
          <meta name="robots" content="index, follow" />
          <meta name="viewport" content="width=device-width, initial-scale=1" />
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
