const { PHASE_DEVELOPMENT_SERVER } = require("next/constants");

module.exports = (phase, { defaultConfig }) => {
  if (phase === PHASE_DEVELOPMENT_SERVER) {
    // development config
    return {
      env: {
        BENEATH_ENV: "development",
      },
    };
  }

  // production config
  return {
    env: {
      BENEATH_ENV: "production",
    }
  };
};
