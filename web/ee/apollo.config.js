module.exports = {
  client: {
    includes: ["apollo/**/*.ts"],
    excludes: ["**/node_modules", "**/__tests__", ".next", "kube", "ee/apollo/types"],
    service: {
      name: "beneath-control-ee",
      url: "http://localhost:4000/ee/graphql",
    },
  },
};
