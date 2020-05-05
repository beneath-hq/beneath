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
        bio
        photoURL
        readQuota
        writeQuota
      }
      services {
        serviceID
        name
        kind
      }
      projects {
        projectID
        name
        displayName
        description
        photoURL
        public
      }
      personal
    }
  }
`;

export const QUERY_USERS_ORGANIZATION_PERMISSIONS = gql`
  query UsersOrganizationPermissions($organizationID: UUID!){
    usersOrganizationPermissions(organizationID: $organizationID) {
      userID
      organizationID
    	view
    	admin
    }
  }
`;

export const ADD_USER_TO_ORGANIZATION = gql`
  mutation InviteUserToOrganization($userID: UUID!, $organizationID: UUID!, $view: Boolean!, $create: Boolean!, $admin: Boolean!) {
    inviteUserToOrganization(userID: $userID, organizationID: $organizationID, view: $view, create: $create, admin: $admin)
  }
`;
