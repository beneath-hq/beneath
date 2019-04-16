module.exports = {
  "port": 5432,
  "type": "postgres",
  "host": "localhost",
  "database": "benjamin",
  "username": "benjamin",
  "password": "",
  "synchronize": true,
  "logging": ["error"],
  // "logger": , // try out with winston
  "cache": !process.env.REDIS_URL ? undefined : {
    "options": {
      "url": process.env.REDIS_URL,
    },
    "type": "redis",
  },
  "entities": [
    "src/entities/**/*.ts",
  ],
  "migrations": [
    "src/migration/**/*.ts",
  ],
  "subscribers": [
    "src/subscriber/**/*.ts",
  ],
  "cli": {
    "entitiesDir": "src/entities",
    "migrationsDir": "src/migration",
    "subscribersDir": "src/subscriber",
  },
};
