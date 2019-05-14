import gql from "graphql-tag";

export const QUERY_USER = gql`
  query User($userId: ID!) {
    user(userId: $userId) {
      userId
      name
      bio
      photoUrl
      createdOn
      projects {
        projectId
        name
        displayName
        description
        photoUrl
      }
    }
  }
`;

export const QUERY_ME = gql`
  query {
    me {
      userId
      user {
        userId
        username
        name
        bio
        photoUrl
        createdOn
      }
      email
      updatedOn
    }
  }
`;

export const UPDATE_ME = gql`
  mutation UpdateMe($name: String, $bio: String) {
    updateMe(name: $name, bio: $bio) {
      userId
      user {
        userId
        name
        bio
      }
      updatedOn
    }
  }
`;
