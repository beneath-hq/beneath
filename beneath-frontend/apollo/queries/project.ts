import gql from "graphql-tag";

export const EXPLORE_PROJECTS = gql`
  query ExploreProjects {
    exploreProjects {
      projectID
      name
      displayName
      description
      photoURL
      createdOn
      updatedOn
    }
  }
`;

export const QUERY_PROJECT = gql`
  query ProjectByName($name: String!) {
    projectByName(name: $name) {
      projectID
      name
      displayName
      site
      description
      photoURL
      createdOn
      updatedOn
      users {
        userID
        name
        username
        photoURL
      }
      streams {
        streamID
        name
        description
        external
      }
    }
  }
`;

export const NEW_PROJECT = gql`
  mutation CreateProject($name: String!, $displayName: String!, $site: String, $description: String, $photoURL: String) {
    createProject(name: $name, displayName: $displayName, site: $site, description: $description, photoURL: $photoURL) {
      projectID
      name
      displayName
      site
      description
      photoURL
      createdOn
      updatedOn
      users {
        userID
        name
        username
        photoURL
      }
    }
  }
`;

export const UPDATE_PROJECT = gql`
  mutation UpdateProject($projectID: UUID!, $displayName: String, $site: String, $description: String, $photoURL: String) {
    updateProject(projectID: $projectID, displayName: $displayName, site: $site, description: $description, photoURL: $photoURL) {
      projectID
      displayName
      site
      description
      photoURL
      updatedOn
    }
  }
`;

export const ADD_MEMBER = gql`
  mutation AddUserToProject($email: String!, $projectID: UUID!) {
    addUserToProject(email: $email, projectID: $projectID) {
      userID
      name
      username
      photoURL
    }
  }
`;

export const REMOVE_MEMBER = gql`
  mutation RemoveUserFromProject($userID: UUID!, $projectID: UUID!) {
    removeUserFromProject(userID: $userID, projectID: $projectID)
  }
`;
