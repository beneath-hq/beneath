import gql from "graphql-tag";

export const QUERY_KEYS = gql`
  query Keys($userId: ID, $projectId: ID) {
    keys(userId: $userId, projectId: $projectId) {
      keyId
      description
      prefix
      role
      createdOn
    }
  }
`;

export const ISSUE_KEY = gql`
  mutation IssueKey($userId: ID, $projectId: ID, $readonly: Boolean!, $description: String) {
    issueKey(userId: $userId, projectId: $projectId, readonly: $readonly, description: $description) {
      keyString
      key {
        keyId
        description
        prefix
        role
        createdOn
      }
    }
  }
`;

export const REVOKE_KEY = gql`
  mutation RevokeKey($keyId: ID!) {
    revokeKey(keyId: $keyId)
  }
`;
