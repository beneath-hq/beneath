from beneath.connection import Connection
from beneath.utils import format_entity_name


class Projects:

  def __init__(self, conn: Connection):
    self.conn = conn

  async def find_by_name(self, name):
    result = await self.conn.query_control(
      variables={
        'name': format_entity_name(name),
      },
      query="""
        query ProjectByName($name: String!) {
          projectByName(name: $name) {
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
            users {
              username
            }
          }
        }
      """
    )
    return result['projectByName']

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
            users {
              userID
              name
              username
              photoURL
            }
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

  async def update_organization(self, project_id, organization_id):
    result = await self.conn.query_control(
      variables={
        'projectID': project_id,
        'organizationID': organization_id,
      },
      query="""
        mutation UpdateProjectOrganization($projectID: UUID!, $organizationID: UUID!) {
          updateProjectOrganization(projectID: $projectID, organizationID: $organizationID) {
            projectID
            displayName
            public
            site
            description
            photoURL
            updatedOn
            organization {
              name
            }
          }
        }
      """
    )
    return result['updateProjectOrganization']

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

  async def add_user(self, project_id, username, view, create, admin):
    result = await self.conn.query_control(
      variables={
        'username': username,
        'projectID': project_id,
        'view': view,
        'create': create,
        'admin': admin,
      },
      query="""
        mutation AddUserToProject($username: String!, $projectID: UUID!, $view: Boolean!, $create: Boolean!, $admin: Boolean!) {
          addUserToProject(username: $username, projectID: $projectID, view: $view, create: $create, admin: $admin) {
            userID
            username
            name
            createdOn
            projects {
              name
            }
            readQuota
            writeQuota
          }
        }
      """
    )
    return result['addUserToProject']

  async def remove_user(self, project_id, user_id):
    result = await self.conn.query_control(
      variables={
        'userID': user_id,
        'projectID': project_id,
      },
      query="""
        mutation RemoveUserFromProject($userID: UUID!, $projectID: UUID!) {
          removeUserFromProject(userID: $userID, projectID: $projectID)
        }
      """
    )
    return result['removeUserFromProject']
