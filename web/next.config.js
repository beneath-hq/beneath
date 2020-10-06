const withCSS = require("@zeit/next-css");
const MonacoWebpackPlugin = require("monaco-editor-webpack-plugin");
const { PHASE_DEVELOPMENT_SERVER } = require("next/constants");

module.exports = (phase) => {
  // setup monaco editor
  const config = withCSS({
    webpack: (config) => {
      config.module.rules.push({
        test: /\.(png|jpg|gif|svg|eot|ttf|woff|woff2)$/,
        use: {
          loader: "url-loader",
          options: {
            limit: 100000,
          },
        },
      });

      config.plugins.push(
        new MonacoWebpackPlugin({
          // Add languages as needed...
          languages: ["json", "graphql", "sql"],
          filename: "static/[name].worker.js",
        })
      );

      return config;
    },
  });

  // add BENEATH_ENV variable
  if (phase === PHASE_DEVELOPMENT_SERVER) {
    config.env = Object.assign({}, config.env, {
      BENEATH_ENV: "development",
    });
  } else {
    config.env = Object.assign({}, config.env, {
      BENEATH_ENV: "production",
    });
  }

  // return
  return config;
};
