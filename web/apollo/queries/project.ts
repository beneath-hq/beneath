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
        external
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

export const UPDATE_PROJECT = gql`
  mutation UpdateProject($projectID: UUID!, $displayName: String, $site: String, $description: String, $photoURL: String) {
    updateProject(projectID: $projectID, displayName: $displayName, site: $site, description: $description, photoURL: $photoURL) {
      projectID
      displayName
      site
      description
      photoURL
      updatedOn
    }
  }
`;
