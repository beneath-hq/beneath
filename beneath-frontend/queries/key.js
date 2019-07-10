import gql from "graphql-tag";

export const QUERY_USER_KEYS = gql`
  query KeysForUser($userID: UUID!) {
    keysForUser(userID: $userID) {
      keyID
      description
      prefix
      role
      createdOn
    }
  }
`;

export const QUERY_PROJECT_KEYS = gql`
  query KeysForProject($projectID: UUID!) {
    keysForProject(projectID: $projectID) {
      keyID
      description
      prefix
      role
      createdOn
    }
  }
`;

export const ISSUE_USER_KEY = gql`
  mutation IssueUserKey($readonly: Boolean!, $description: String!) {
    issueUserKey(readonly: $readonly, description: $description) {
      keyString
      key {
        keyID
        description
        prefix
        role
        createdOn
      }
    }
  }
`;

export const ISSUE_PROJECT_KEY = gql`
  mutation IssueProjectKey($projectID: UUID!, $readonly: Boolean!, $description: String!) {
    issueProjectKey(projectID: $projectID, readonly: $readonly, description: $description) {
      keyString
      key {
        keyID
        description
        prefix
        role
        createdOn
      }
    }
  }
`;

export const REVOKE_KEY = gql`
  mutation RevokeKey($keyID: UUID!) {
    revokeKey(keyID: $keyID)
  }
`;
