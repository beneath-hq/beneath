import gql from "graphql-tag";

export const REGISTER_USER_CONSENT = gql`
  mutation RegisterUserConsent($userID: UUID!, $terms: Boolean, $newsletter: Boolean) {
    registerUserConsent(userID: $userID, terms: $terms, newsletter: $newsletter) {
      userID
      updatedOn
      consentTerms
      consentNewsletter
    }
  }
`;

export const QUERY_AUTH_TICKET = gql`
  query AuthTicketByID($authTicketID: UUID!) {
    authTicketByID(authTicketID: $authTicketID) {
      authTicketID
      requesterName
      createdOn
      updatedOn
    }
  }
`;

export const UPDATE_AUTH_TICKET = gql`
  mutation UpdateAuthTicket($input: UpdateAuthTicketInput!) {
    updateAuthTicket(input: $input) {
      authTicketID
      updatedOn
    }
  }
`;
