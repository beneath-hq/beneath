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

  async def create(self, name):
    result = await self.conn.query_control(
      variables={
        'name': format_entity_name(name),
      },
      query="""
        mutation CreateOrganization($name: String!) {
          createOrganization(name: $name) {
           	organizationID
            name
            createdOn
            updatedOn
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
    return result['createOrganization']

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

  async def invite_user(self, organization_id, user_id, view, create, admin):
    result = await self.conn.query_control(
      variables={
        'userID': userID,
        'organizationID': organization_id,
        'view': view,
        'create': create,
        'admin': admin,
      },
      query="""
        mutation InviteUserToOrganization($userID: UUID!, $organizationID: UUID!, $view: Boolean!, $create: Boolean!, $admin: Boolean!) {
          inviteUserToOrganization(userID: $userID, organizationID: $organizationID, view: $view, create: $create, admin: $admin)
        }
      """
    )
    return result['inviteUserToOrganization']

  async def accept_invite(self, organization_id):
    result = await self.conn.query_control(
      variables={
        'organizationID': organization_id,
      },
      query="""
        mutation AcceptOrganizationInvite($organizationID: UUID!) {
          acceptOrganizationInvite(organizationID: $organizationID) 
        }
      """
    )
    return result['acceptOrganizationInvite']

  async def update_user_quota(self, user_id, read_quota_bytes, write_quota_bytes):
    result = await self.conn.query_control(
      variables={
        'userID': user_id,
        'readQuota': read_quota_bytes,
        'writeQuota': write_quota_bytes,
      },
      query="""
        mutation UpdateUserQuotas($userID: UUID!, $readQuota: Int, $writeQuota: Int) {
          updateUserQuotas(userID: $userID, readQuota: $readQuota, writeQuota: $writeQuota) {
            userID
            username
            readQuota
            writeQuota
          }
        }
      """
    )
    return result['updateUserQuotas']

  async def leave(self, user_id):
    result = await self.conn.query_control(
      variables={
        'userID': user_id,
      },
      query="""
        mutation LeaveBillingOrganization($userID: UUID!) {
          leaveBillingOrganization(userID: $userID) {
            userID
          }
        }
      """
    )
    return result['leaveBillingOrganization']

  async def transfer_project(self, project_id, new_organization_id):
    result = await self.conn.query_control(
      variables={
        'projectID': project_id,
        'organizationID': new_organization_id,
      },
      query="""
        mutation TransferProjectToOrganization($projectID: UUID!, $organizationID: UUID!) {
          transferProjectToOrganization(projectID: $projectID, organizationID: $organizationID) {
            projectID
            organization {
              organizationID
              name
            }
          }
        }
      """
    )
    return result['transferProjectToOrganization']

  async def transfer_service(self, service_id, new_organization_id):
    result = await self.conn.query_control(
      variables={
        'serviceID': service_id,
        'organizationID': new_organization_id,
      },
      query="""
        mutation TransferServiceToOrganization($serviceID: UUID!, $organizationID: UUID!) {
          transferServiceToOrganization(serviceID: $serviceID, organizationID: $organizationID) {
            serviceID
            organization {
              organizationID
              name
            }
          }
        }
      """
    )
    return result['transferServiceToOrganization']
