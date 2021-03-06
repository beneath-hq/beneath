import gql from "graphql-tag";

export const QUERY_ME = gql`
  query Me {
    me {
      organizationID
      name
      displayName
      description
      photoURL
      createdOn
      projects {
        projectID
        name
        displayName
        description
        photoURL
      }
      personalUserID
      updatedOn
      quotaStartTime
      quotaEndTime
      prepaidReadQuota
      prepaidWriteQuota
      prepaidScanQuota
      readQuota
      writeQuota
      scanQuota
      readUsage
      writeUsage
      scanUsage
      personalUser {
        userID
        email
        consentTerms
        consentNewsletter
        createdOn
        updatedOn
        quotaStartTime
        quotaEndTime
        readQuota
        writeQuota
        scanQuota
        billingOrganizationID
        billingOrganization {
          organizationID
          name
          displayName
        }
      }
      permissions {
        view
        create
        admin
      }
    }
  }
`;

export const QUERY_ORGANIZATION = gql`
  query OrganizationByName($name: String!) {
    organizationByName(name: $name) {
      organizationID
      name
      displayName
      description
      photoURL
      createdOn
      projects {
        projectID
        name
        displayName
        description
        photoURL
        public
      }
      personalUserID
      ... on PrivateOrganization {
        updatedOn
        quotaStartTime
        quotaEndTime
        prepaidReadQuota
        prepaidWriteQuota
        prepaidScanQuota
        readQuota
        writeQuota
        scanQuota
        readUsage
        writeUsage
        scanUsage
        personalUser {
          userID
          email
          consentTerms
          consentNewsletter
          createdOn
          updatedOn
          quotaStartTime
          quotaEndTime
          readQuota
          writeQuota
          scanQuota
          billingOrganizationID
          billingOrganization {
            organizationID
            name
            displayName
          }
        }
        permissions {
          view
          create
          admin
        }
      }
    }
  }
`;

export const QUERY_ORGANIZATION_MEMBERS = gql`
  query OrganizationMembers($organizationID: UUID!) {
    organizationMembers(organizationID: $organizationID) {
      organizationID
      userID
      billingOrganizationID
      name
      displayName
      photoURL
      view
      create
      admin
      readQuota
      writeQuota
      scanQuota
    }
  }
`;

export const UPDATE_ORGANIZATION = gql`
  mutation UpdateOrganization(
    $organizationID: UUID!
    $name: String
    $displayName: String
    $description: String
    $photoURL: String
  ) {
    updateOrganization(
      organizationID: $organizationID
      name: $name
      displayName: $displayName
      description: $description
      photoURL: $photoURL
    ) {
      organizationID
      name
      displayName
      description
      photoURL
      createdOn
      updatedOn
      personalUserID
      personalUser {
        userID
        email
        createdOn
        updatedOn
        readQuota
        writeQuota
        scanQuota
        billingOrganizationID
      }
    }
  }
`;

export const UPDATE_USER_ORGANIZATION_PERMISSIONS = gql`
  mutation UpdateUserOrganizationPermissions($userID: UUID!, $organizationID: UUID!, $view: Boolean, $create: Boolean, $admin: Boolean) {
    updateUserOrganizationPermissions(userID: $userID, organizationID: $organizationID, view: $view, create: $create, admin: $admin) {
      userID
      organizationID
      view
      create
      admin
    }
  }
`;
