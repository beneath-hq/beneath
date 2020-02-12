import gql from "graphql-tag";

export const QUERY_USER_SECRETS = gql`
  query SecretsForUser($userID: UUID!) {
    secretsForUser(userID: $userID) {
      userSecretID
      description
      prefix
      createdOn
      readOnly
      publicOnly
    }
  }
`;

export const ISSUE_USER_SECRET = gql`
  mutation IssueUserSecret($description: String!, $readOnly: Boolean!, $publicOnly: Boolean!) {
    issueUserSecret(description: $description, readOnly: $readOnly, publicOnly: $publicOnly) {
      token
      secret {
        userSecretID
        description
        prefix
        createdOn
        readOnly
        publicOnly
      }
    }
  }
`;

export const REVOKE_USER_SECRET = gql`
  mutation RevokeUserSecret($secretID: UUID!) {
    revokeUserSecret(secretID: $secretID)
  }
`;
