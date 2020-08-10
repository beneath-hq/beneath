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
      }
      services {
        serviceID
        name
        description
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

export const STAGE_PROJECT = gql`
  mutation StageProject($organizationName: String!, $projectName: String!, $displayName: String, $public: Boolean, $description: String, $site: String, $photoURL: String) {
    stageProject(organizationName: $organizationName, projectName: $projectName, displayName: $displayName, public: $public, description: $description, site: $site, photoURL: $photoURL) {
      projectID
      name
      displayName
      public
      description
      site
      photoURL
      updatedOn
    }
  }
`;
