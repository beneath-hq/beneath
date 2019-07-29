module.exports = {
  client: {
    includes: ['queries/**/*.ts'],
    excludes: ['**/node_modules', '**/__tests__', '.next', 'kube'], 
    service: {
      name: 'beneath-control',
      url: 'http://localhost:4000/graphql',
    }
  }
};
