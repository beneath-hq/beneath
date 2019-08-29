import gql from "graphql-tag";

export const QUERY_USER_SECRETS = gql`
  query SecretsForUser($userID: UUID!) {
    secretsForUser(userID: $userID) {
      secretID
      description
      prefix
      role
      createdOn
    }
  }
`;

export const QUERY_PROJECT_SECRETS = gql`
  query SecretsForProject($projectID: UUID!) {
    secretsForProject(projectID: $projectID) {
      secretID
      description
      prefix
      role
      createdOn
    }
  }
`;

export const ISSUE_USER_SECRET = gql`
  mutation IssueUserSecret($readonly: Boolean!, $description: String!) {
    issueUserSecret(readonly: $readonly, description: $description) {
      secretString
      secret {
        secretID
        description
        prefix
        role
        createdOn
      }
    }
  }
`;

export const ISSUE_PROJECT_SECRET = gql`
  mutation IssueProjectSecret($projectID: UUID!, $readonly: Boolean!, $description: String!) {
    issueProjectSecret(projectID: $projectID, readonly: $readonly, description: $description) {
      secretString
      secret {
        secretID
        description
        prefix
        role
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
