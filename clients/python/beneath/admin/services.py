from beneath.connection import Connection
from beneath.utils import format_entity_name


class Services:

  def __init__(self, conn: Connection):
    self.conn = conn

  async def find_by_organization_and_name(self, organization_name, name):
    result = await self.conn.query_control(
      variables={
        'name': format_entity_name(name),
        'organizationName': format_entity_name(organization_name),
      },
      query="""
        query ServiceByNameAndOrganization($name: String!, $organizationName: String!) {
          serviceByNameAndOrganization(name: $name, organizationName: $organizationName) {
            serviceID
            name
            kind
            readQuota
            writeQuota
          }
        }
      """
    )
    return result['serviceByNameAndOrganization']

  async def create(self, name, organization_id, read_quota_bytes, write_quota_bytes):
    result = await self.conn.query_control(
      variables={
        'name': format_entity_name(name),
        'organizationID': organization_id,
        'readQuota': read_quota_bytes,
        'writeQuota': write_quota_bytes,
      },
      query="""
        mutation CreateService($name: String!, $organizationID: UUID!, $readQuota: Int!, $writeQuota: Int!) {
          createService(name: $name, organizationID: $organizationID, readQuota: $readQuota, writeQuota: $writeQuota) {
            serviceID
            name
            kind
            readQuota
            writeQuota
          }
        }
      """
    )
    return result['createService']


  async def update_details(self, service_id, name):
    result = await self.conn.query_control(
      variables={
        'serviceID': service_id,
        'name': format_entity_name(name) if name is not None else None,
      },
      query="""
        mutation UpdateService($serviceID: UUID!, $name: String) {
          updateService(serviceID: $serviceID, name: $name) {
            serviceID
            name
            kind
            readQuota
            writeQuota
          }
        }
      """
    )
    return result['updateService']

  async def update_quota(self, service_id, read_quota_bytes, write_quota_bytes):
    result = await self.conn.query_control(
      variables={
        'serviceID': service_id,
        'readQuota': read_quota_bytes,
        'writeQuota': write_quota_bytes,
      },
      query="""
        mutation UpdateServiceQuotas($serviceID: UUID!, $readQuota: Int, $writeQuota: Int) {
          updateServiceQuotas(serviceID: $serviceID, readQuota: $readQuota, writeQuota: $writeQuota) {
            serviceID
            name
            kind
            readQuota
            writeQuota
          }
        }
      """
    )
    return result['updateServiceQuotas']

  async def update_permissions_for_stream(self, service_id, stream_id, read, write):
    result = await self.conn.query_control(
      variables={
        'serviceID': service_id,
        'streamID': stream_id,
        'read': read,
        'write': write,
      },
      query="""
        mutation UpdateServicePermissions($serviceID: UUID!, $streamID: UUID!, $read: Boolean, $write: Boolean) {
          updateServiceStreamPermissions(serviceID: $serviceID, streamID: $streamID, read: $read, write: $write) {
            serviceID
            streamID
            read
            write
          }
        }
      """
    )
    return result['updateServiceStreamPermissions']

  async def delete(self, service_id):
    result = await self.conn.query_control(
      variables={
        'serviceID': service_id,
      },
      query="""
        mutation DeleteService($serviceID: UUID!) {
          deleteService(serviceID: $serviceID)
        }
      """
    )
    return result['deleteService']

  async def issue_secret(self, service_id, description):
    result = await self.conn.query_control(
      variables={
        'serviceID': service_id,
        'description': description,
      },
      query="""
        mutation IssueServiceSecret($serviceID: UUID!, $description: String!) {
          issueServiceSecret(serviceID: $serviceID, description: $description) {
            token
          }
        }
      """
    )
    return result['issueServiceSecret']

  async def list_secrets(self, service_id):
    result = await self.conn.query_control(
      variables={
        'serviceID': service_id,
      },
      query="""
        query SecretsForService($serviceID: UUID!) {
          secretsForService(serviceID: $serviceID) {
            secretID
            description
            prefix
            createdOn
            updatedOn
          }
        }
      """
    )
    return result['secretsForService']
