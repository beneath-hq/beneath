import gql from "graphql-tag";

export const EXPLORE_PROJECTS = gql`
  query ExploreProjects {
    exploreProjects {
      projectID
      name
      displayName
      description
      photoURL
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
      streams {
        streamID
        name
        description
        createdOn
        instancesCreatedCount
        instancesDeletedCount
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
