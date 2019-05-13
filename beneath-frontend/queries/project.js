import gql from "graphql-tag";

export const QUERY_PROJECT = gql`
  query Project($name: String) {
    project(name: $name) {
      projectId
      name
      displayName
      site
      description
      createdOn
      updatedOn
      users {
        userId
        name
        username
        photoUrl
      }
    }
  }
`;

export const REMOVE_MEMBER = gql`
  mutation RemoveUserFromProject($userId: ID!, $projectId: ID!) {
    removeUserFromProject(userId: $userId, projectId: $projectId)
  }
`;
