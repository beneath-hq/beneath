module.exports = {
  "type": "postgres",
  "host": "localhost",
  "port": 5432,
  "database": "benjamin",
  "username": "benjamin",
  "password": "",
  "synchronize": true,
  "logging": ["error"],
  // "logger": , // try out with winston
  "cache": !process.env.REDIS_URL ? undefined : {
    "type": "redis",
    "options": {
      "url": process.env.REDIS_URL,
    }
  },
  "entities": [
    "src/entities/**/*.ts"
  ],
  "migrations": [
    "src/migration/**/*.ts"
  ],
  "subscribers": [
    "src/subscriber/**/*.ts"
  ],
  "cli": {
    "entitiesDir": "src/entities",
    "migrationsDir": "src/migration",
    "subscribersDir": "src/subscriber"
  }
};