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
      readQuota
      writeQuota
      readUsage
      writeUsage
      services {
        serviceID
        name
        kind
      }
      personalUser {
        userID
        email
        createdOn
        updatedOn
        readQuota
        writeQuota
        billingOrganizationID
        billingOrganization {
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
      }
      personalUserID
      ... on PrivateOrganization {
        updatedOn
        prepaidReadQuota
        prepaidWriteQuota
        readQuota
        writeQuota
        readUsage
        writeUsage
        services {
          serviceID
          name
          kind
        }
        personalUser {
          userID
          email
          createdOn
          updatedOn
          readQuota
          writeQuota
          billingOrganizationID
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
  query OrganizationMembers($organizationID: UUID!){
    organizationMembers(organizationID: $organizationID) {
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
    }
  }
`;

export const UPDATE_ORGANIZATION = gql`
  mutation UpdateOrganization($organizationID: UUID!, $name: String, $displayName: String, $description: String, $photoURL: String) {
    updateOrganization(organizationID: $organizationID, name: $name, displayName: $displayName, description: $description, photoURL: $photoURL) {
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
        billingOrganizationID
      }
    }
  }
`;
