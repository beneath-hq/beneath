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
