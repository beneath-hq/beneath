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

export const QUERY_ME = gql`
  query {
    me {
      userID
      email
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
    }
  }
`;

export const UPDATE_ME = gql`
  mutation UpdateMe($name: String, $bio: String) {
    updateMe(name: $name, bio: $bio) {
      userID
      user {
        userID
        name
        bio
      }
      updatedOn
    }
  }
`;
