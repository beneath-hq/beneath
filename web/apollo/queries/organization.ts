import gql from "graphql-tag";

export const QUERY_ORGANIZATION = gql`
  query OrganizationByName($name: String!) {
    organizationByName(name: $name) {
      organizationID
      name
      createdOn
      updatedOn
      users {
        userID
        name
        username
        photoURL
        readQuota
        writeQuota
      }
      services {
        serviceID
        name
        kind
      }
    }
  }
`;

export const QUERY_USERS_ORGANIZATION_PERMISSIONS = gql`
  query UsersOrganizationPermissions($organizationID: UUID!){
    usersOrganizationPermissions(organizationID: $organizationID) {
    	user {
        userID
      }
    	organization {
        organizationID
      }
    	view
    	admin
    }
  }
`;

export const ADD_USER_TO_ORGANIZATION = gql`
  mutation InviteUserToOrganization($username: String!, $organizationID: UUID!, $view: Boolean!, $admin: Boolean!) {
    inviteUserToOrganization(username: $username, organizationID: $organizationID, view: $view, admin: $admin) {
      userID
    }
  }
`
