import gql from "graphql-tag";

export const EXPLORE_PROJECTS = gql`
  query ExploreProjects {
    exploreProjects {
      projectID
      name
      displayName
      description
      photoURL
      public
      createdOn
      updatedOn
      organization {
        organizationID
        name
      }
    }
  }
`;

export const QUERY_PROJECTS_FOR_USER = gql`
  query ProjectsForUser($userID: UUID!) {
    projectsForUser(userID: $userID) {
      projectID
      name
      displayName
      description
      photoURL
      public
      createdOn
      updatedOn
      organization {
        organizationID
        name
      }
    }
  }
`;

export const QUERY_PROJECT = gql`
  query ProjectByOrganizationAndName($organizationName: String!, $projectName: String!) {
    projectByOrganizationAndName(organizationName: $organizationName, projectName: $projectName) {
      projectID
      name
      displayName
      site
      description
      photoURL
      public
      createdOn
      updatedOn
      organization {
        organizationID
        name
      }
      tables {
        tableID
        name
        description
        createdOn
        meta
        instancesCreatedCount
        instancesDeletedCount
        primaryTableInstanceID
      }
      services {
        serviceID
        name
        description
        createdOn
      }
      permissions {
        view
        create
        admin
      }
    }
  }
`;

export const QUERY_PROJECT_MEMBERS = gql`
  query ProjectMembers($projectID: UUID!) {
    projectMembers(projectID: $projectID) {
      projectID
      userID
      name
      displayName
      photoURL
      view
      create
      admin
    }
  }
`;

export const CREATE_PROJECT = gql`
  mutation CreateProject($input: CreateProjectInput!) {
    createProject(input: $input) {
      projectID
      name
      displayName
      public
      description
      site
      photoURL
      updatedOn
      organization {
        organizationID
        name
      }
    }
  }
`;

export const UPDATE_PROJECT = gql`
  mutation UpdateProject($input: UpdateProjectInput!) {
    updateProject(input: $input) {
      projectID
      name
      displayName
      public
      description
      site
      photoURL
      updatedOn
      organization {
        organizationID
        name
      }
    }
  }
`;

export const UPDATE_USER_PROJECT_PERMISSIONS = gql`
  mutation UpdateUserProjectPermissions($userID: UUID!, $projectID: UUID!, $view: Boolean, $create: Boolean, $admin: Boolean) {
    updateUserProjectPermissions(userID: $userID, projectID: $projectID, view: $view, create: $create, admin: $admin) {
      userID
      projectID
      view
      create
      admin
    }
  }
`;
