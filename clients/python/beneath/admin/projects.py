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
          }
        }
      """
    )
    return result['projectByOrganizationAndName']

  async def create(
    self,
    name,
    organization_id,
    public,
    display_name=None,
    description=None,
    site_url=None,
    photo_url=None,
  ):
    result = await self.conn.query_control(
      variables={
        'name': format_entity_name(name),
        'displayName': display_name,
        'organizationID': organization_id,
        'public': public,
        'description': description,
        'site': site_url,
        'photoURL': photo_url,
      },
      query="""
        mutation CreateProject($name: String!, $displayName: String, $organizationID: UUID!, $public: Boolean! $site: String, $description: String, $photoURL: String) {
          createProject(name: $name, displayName: $displayName, organizationID: $organizationID, public: $public, site: $site, description: $description, photoURL: $photoURL) {
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
    return result['createProject']

  async def update_details(
    self,
    project_id,
    display_name,
    public,
    description=None,
    site_url=None,
    photo_url=None,
  ):
    result = await self.conn.query_control(
      variables={
        'projectID': project_id,
        'displayName': display_name,
        'public': public,
        'description': description,
        'site': site_url,
        'photoURL': photo_url,
      },
      query="""
        mutation UpdateProject($projectID: UUID!, $displayName: String, $public: Boolean, $site: String, $description: String, $photoURL: String) {
          updateProject(projectID: $projectID, displayName: $displayName, public: $public, site: $site, description: $description, photoURL: $photoURL) {
            projectID
            displayName
            public
            site
            description
            photoURL
            updatedOn
          }
        }
      """
    )
    return result['updateProject']

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
