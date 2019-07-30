module.exports = {
  client: {
    includes: ["apollo/**/*.ts"],
    excludes: ["**/node_modules", "**/__tests__", ".next", "kube", "apollo/types"],
    service: {
      name: "beneath-control",
      url: "http://localhost:4000/graphql",
    },
  },
};
