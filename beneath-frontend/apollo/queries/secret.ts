import gql from "graphql-tag";

export const QUERY_USER_SECRETS = gql`
  query SecretsForUser($userID: UUID!) {
    secretsForUser(userID: $userID) {
      secretID
      description
      prefix
      createdOn
    }
  }
`;

export const ISSUE_USER_SECRET = gql`
  mutation IssueUserSecret($description: String!) {
    issueUserSecret(description: $description) {
      secretString
      secret {
        secretID
        description
        prefix
        createdOn
      }
    }
  }
`;

export const REVOKE_SECRET = gql`
  mutation RevokeSecret($secretID: UUID!) {
    revokeSecret(secretID: $secretID)
  }
`;
