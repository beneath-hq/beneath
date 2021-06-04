const { PHASE_DEVELOPMENT_SERVER } = require("next/constants");

module.exports = (phase) => {
  // setup monaco editor
  const config = {};

  // add BENEATH_ENV variable
  if (phase === PHASE_DEVELOPMENT_SERVER) {
    config.env = Object.assign({}, config.env, {
      BENEATH_EE: process.env.BENEATH_EE,
      BENEATH_ENV: "development",
    });
  } else {
    config.env = Object.assign({}, config.env, {
      BENEATH_EE: process.env.BENEATH_EE,
      BENEATH_ENV: "production",
    });
  }

  // use webpack 5
  config.future = {
    webpack5: true,
  };

  // return
  return config;
};
