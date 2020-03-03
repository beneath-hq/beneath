from beneath.connection import Connection
from beneath.utils import format_entity_name


class Organizations:

  def __init__(self, conn: Connection):
    self.conn = conn

  async def find_by_name(self, name):
    result = await self.conn.query_control(
      variables={
        'name': format_entity_name(name),
      },
      query="""
        query OrganizationByName($name: String!) {
          organizationByName(name: $name) {
            organizationID
            name
            createdOn
            updatedOn
            services {
              serviceID
              name
              kind
              readQuota
              writeQuota
            }
            users {
              userID
              username
              name
              readQuota
              writeQuota
            }
          }
        }
      """
    )
    return result['organizationByName']

  async def update_name(self, organization_id, name):
    result = await self.conn.query_control(
      variables={
        'organizationID': organization_id,
        'name': format_entity_name(name),
      },
      query="""
        mutation UpdateOrganizationName($organizationID: UUID!, $name: String!) {
          updateOrganizationName(organizationID: $organizationID, name: $name) {
           	organizationID
            name
            createdOn
            updatedOn
            services {
              serviceID
              name
              kind
              readQuota
              writeQuota
            }
            users {
              userID
              username
              name
              readQuota
              writeQuota
            }
          }
        }
      """
    )
    return result['updateOrganizationName']

  async def get_member_permissions(self, organization_id):
    result = await self.conn.query_control(
      variables={
        'organizationID': organization_id,
      },
      query="""
        query UsersOrganizationPermissions($organizationID: UUID!) {
          usersOrganizationPermissions(organizationID: $organizationID) {
            user {
              username
              name
            }
            view
            admin
          }
        }
      """
    )
    return result['usersOrganizationPermissions']

  async def add_user(self, organization_id, username, view, admin):
    result = await self.conn.query_control(
      variables={
        'username': username,
        'organizationID': organization_id,
        'view': view,
        'admin': admin,
      },
      query="""
        mutation InviteUserToOrganization($username: String!, $organizationID: UUID!, $view: Boolean!, $admin: Boolean!) {
          inviteUserToOrganization(username: $username, organizationID: $organizationID, view: $view, admin: $admin) {
            userID
            username
            name
            createdOn
            projects {
              name
              public
            }
            readQuota
            writeQuota
          }
        }
      """
    )
    return result['inviteUserToOrganization']

  async def join(self, name):
    result = await self.conn.query_control(
      variables={
        'organizationName': format_entity_name(name),
      },
      query="""
        mutation JoinOrganization($organizationName: String!) {
          joinOrganization(organizationName: $organizationName) {
            user {
              username
            }
            organization {
              name
            }
            readQuota
            writeQuota
          }
        }
      """
    )
    return result['joinOrganization']

  async def remove_user(self, organization_id, user_id):
    result = await self.conn.query_control(
      variables={
        'userID': user_id,
        'organizationID': organization_id,
      },
      query="""
        mutation RemoveUserFromOrganization($userID: UUID!, $organizationID: UUID!) {
          removeUserFromOrganization(userID: $userID, organizationID: $organizationID)
        }
      """
    )
    return result['removeUserFromOrganization']

  async def update_permissions_for_user(self, organization_id, user_id, view, admin):
    result = await self.conn.query_control(
      variables={
        'userID': user_id,
        'organizationID': organization_id,
        'view': view,
        'admin': admin,
      },
      query="""
        mutation UpdateUserOrganizationPermissions($userID: UUID!, $organizationID: UUID!, $view: Boolean, $admin: Boolean) {
          updateUserOrganizationPermissions(userID: $userID, organizationID: $organizationID, view: $view, admin: $admin) {
            user {
              username
            }
            organization {
              name
            }
            view
            admin
          }
        }
      """
    )
    return result['updateUserOrganizationPermissions']

  async def update_quotas_for_user(self, organization_id, user_id, read_quota_bytes, write_quota_bytes):
    result = await self.conn.query_control(
      variables={
        'userID': user_id,
        'organizationID': organization_id,
        'readQuota': read_quota_bytes,
        'writeQuota': write_quota_bytes,
      },
      query="""
        mutation UpdateUserOrganizationQuotas($userID: UUID!, $organizationID: UUID!, $readQuota: Int, $writeQuota: Int) {
          updateUserOrganizationQuotas(userID: $userID, organizationID: $organizationID, readQuota: $readQuota, writeQuota: $writeQuota) {
            username
            readQuota
            writeQuota
          }
        }
      """
    )
    return result['updateUserOrganizationQuotas']
