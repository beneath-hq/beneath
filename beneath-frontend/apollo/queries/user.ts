import gql from "graphql-tag";

export const QUERY_USER = gql`
  query User($userID: UUID!) {
    user(userID: $userID) {
      userID
      username
      name
      bio
      photoURL
      createdOn
      projects {
        projectID
        name
        displayName
        description
        photoURL
      }
    }
  }
`;

export const QUERY_USER_BY_USERNAME = gql`
  query UserByUsername($username: String!) {
    userByUsername(username: $username) {
      userID
      username
      name
      bio
      photoURL
      createdOn
      projects {
        projectID
        name
        displayName
        description
        photoURL
      }
    }
  }
`;

export const QUERY_ME = gql`
  query Me {
    me {
      userID
      email
      readUsage
      readQuota
      writeUsage
      writeQuota
      updatedOn
      user {
        userID
        username
        name
        bio
        photoURL
        createdOn
        projects {
          projectID
          name
          displayName
          description
          photoURL
        }
      }
      organization {
        organizationID
        name
        personal
      }      
    }
  }
`;

export const UPDATE_ME = gql`
  mutation UpdateMe($username: String, $name: String, $bio: String) {
    updateMe(username: $username, name: $name, bio: $bio) {
      userID
      readUsage
      readQuota
      writeUsage
      writeQuota
      updatedOn
      user {
        userID
        username
        name
        bio
        photoURL
      }
    }
  }
`;
