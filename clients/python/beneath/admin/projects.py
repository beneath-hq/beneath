from beneath.connection import Connection
from beneath.utils import format_entity_name


class Projects:

  def __init__(self, conn: Connection):
    self.conn = conn

  async def find_by_organization_and_name(self, organization_name, project_name):
    result = await self.conn.query_control(
      variables={
        'organizationName': format_entity_name(organization_name),
        'projectName': format_entity_name(project_name),
      },
      query="""
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
            streams {
              name
            }
            services {
              name
            }
          }
        }
      """
    )
    return result['projectByOrganizationAndName']

  async def stage(
    self,
    organization_name,
    project_name,
    display_name=None,
    public=None,
    description=None,
    site_url=None,
    photo_url=None,
  ):
    result = await self.conn.query_control(
      variables={
        "organizationName": format_entity_name(organization_name),
        "projectName": format_entity_name(project_name),
        "displayName": display_name,
        "public": public,
        "description": description,
        "site": site_url,
        "photoURL": photo_url,
      },
      query="""
        mutation StageProject(
          $organizationName: String!,
          $projectName: String!,
          $displayName: String,
          $public: Boolean,
          $description: String,
          $site: String,
          $photoURL: String,
        ) {
          stageProject(
            organizationName: $organizationName,
            projectName: $projectName,
            displayName: $displayName,
            public: $public,
            description: $description,
            site: $site,
            photoURL: $photoURL,
          ) {
            projectID
            name
            displayName
            public
            site
            description
            photoURL
            createdOn
            updatedOn
          }
        }
      """
    )
    return result['stageProject']

  async def get_member_permissions(self, project_id):
    result = await self.conn.query_control(
      variables={
        'projectID': project_id,
      },
      query="""
        query ProjectMembers($projectID: UUID!) {
          projectMembers(projectID: $projectID) {
            userID
            name
            displayName
            view
            create
            admin
          }
        }
      """
    )
    return result['projectMembers']

  async def delete(self, project_id):
    result = await self.conn.query_control(
      variables={
        'projectID': project_id,
      },
      query="""
        mutation DeleteProject($projectID: UUID!) {
          deleteProject(projectID: $projectID)
        }
      """
    )
    return result['deleteProject']
