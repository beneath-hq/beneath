import gql from "graphql-tag";

export const QUERY_PROJECT = gql`
  query Project($name: String) {
    project(name: $name) {
      projectId
      name
      displayName
      site
      description
      photoUrl
      createdOn
      updatedOn
      users {
        userId
        name
        username
        photoUrl
      }
      streams {
        streamId
        name
        description
        external
      }
    }
  }
`;

export const NEW_PROJECT = gql`
  mutation CreateProject($name: String!, $displayName: String!, $site: String, $description: String, $photoUrl: String) {
    createProject(name: $name, displayName: $displayName, site: $site, description: $description, photoUrl: $photoUrl) {
      projectId
      name
      displayName
      site
      description
      photoUrl
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

export const UPDATE_PROJECT = gql`
  mutation UpdateProject($projectId: ID!, $displayName: String, $site: String, $description: String, $photoUrl: String) {
    updateProject(projectId: $projectId, displayName: $displayName, site: $site, description: $description, photoUrl: $photoUrl) {
      projectId
      displayName
      site
      description
      photoUrl
      updatedOn
    }
  }
`;

export const ADD_MEMBER = gql`
  mutation AddUserToProject($email: String!, $projectId: ID!) {
    addUserToProject(email: $email, projectId: $projectId) {
      userId
      name
      username
      photoUrl
    }
  }
`;

export const REMOVE_MEMBER = gql`
  mutation RemoveUserFromProject($userId: ID!, $projectId: ID!) {
    removeUserFromProject(userId: $userId, projectId: $projectId)
  }
`;
