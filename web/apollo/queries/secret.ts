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

export const QUERY_SERVICE_SECRETS = gql`
  query SecretsForService($serviceID: UUID!) {
    secretsForService(serviceID: $serviceID) {
      serviceSecretID
      description
      prefix
      createdOn
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

export const ISSUE_SERVICE_SECRET = gql`
  mutation IssueServiceSecret($serviceID: UUID!, $description: String!) {
    issueServiceSecret(serviceID: $serviceID, description: $description) {
      token
      secret {
        serviceSecretID
        description
        prefix
        createdOn
      }
    }
  }
`;

export const REVOKE_USER_SECRET = gql`
  mutation RevokeUserSecret($secretID: UUID!) {
    revokeUserSecret(secretID: $secretID)
  }
`;

export const REVOKE_SERVICE_SECRET = gql`
  mutation RevokeServiceSecret($secretID: UUID!) {
    revokeServiceSecret(secretID: $secretID)
  }
`;
