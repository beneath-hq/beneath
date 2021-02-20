from beneath.admin.base import _ResourceBase
from beneath.utils import format_entity_name


class Services(_ResourceBase):
    async def find_by_organization_project_and_name(
        self,
        organization_name,
        project_name,
        service_name,
    ):
        result = await self.conn.query_control(
            variables={
                "organizationName": format_entity_name(organization_name),
                "projectName": format_entity_name(project_name),
                "serviceName": format_entity_name(service_name),
            },
            query="""
                query ServiceByOrganizationProjectAndName(
                    $organizationName: String!
                    $projectName: String!
                    $serviceName: String!
                ) {
                    serviceByOrganizationProjectAndName(
                        organizationName: $organizationName
                        projectName: $projectName
                        serviceName: $serviceName
                    ) {
                        serviceID
                        name
                        description
                        sourceURL
                        readQuota
                        writeQuota
                        scanQuota
                    }
                }
            """,
        )
        return result["serviceByOrganizationProjectAndName"]

    async def create(
        self,
        organization_name,
        project_name,
        service_name,
        description=None,
        source_url=None,
        read_quota_bytes=None,
        write_quota_bytes=None,
        scan_quota_bytes=None,
        update_if_exists=None,
    ):
        self._before_mutation()
        result = await self.conn.query_control(
            variables={
                "input": {
                    "organizationName": format_entity_name(organization_name),
                    "projectName": format_entity_name(project_name),
                    "serviceName": format_entity_name(service_name),
                    "description": description,
                    "sourceURL": source_url,
                    "readQuota": read_quota_bytes,
                    "writeQuota": write_quota_bytes,
                    "scanQuota": scan_quota_bytes,
                    "updateIfExists": update_if_exists,
                }
            },
            query="""
                mutation CreateService($input: CreateServiceInput!) {
                    createService(input: $input) {
                    serviceID
                    name
                    description
                    sourceURL
                    readQuota
                    writeQuota
                    scanQuota
                    project {
                        projectID
                        name
                        organization {
                        organizationID
                        name
                        }
                    }
                    }
                }
            """,
        )
        return result["createService"]

    async def update(
        self,
        organization_name,
        project_name,
        service_name,
        description=None,
        source_url=None,
        read_quota_bytes=None,
        write_quota_bytes=None,
        scan_quota_bytes=None,
    ):
        self._before_mutation()
        result = await self.conn.query_control(
            variables={
                "input": {
                    "organizationName": format_entity_name(organization_name),
                    "projectName": format_entity_name(project_name),
                    "serviceName": format_entity_name(service_name),
                    "description": description,
                    "sourceURL": source_url,
                    "readQuota": read_quota_bytes,
                    "writeQuota": write_quota_bytes,
                    "scanQuota": scan_quota_bytes,
                }
            },
            query="""
                mutation UpdateService($input: UpdateServiceInput!) {
                    updateService(input: $input) {
                    serviceID
                    name
                    description
                    sourceURL
                    readQuota
                    writeQuota
                    scanQuota
                    project {
                        projectID
                        name
                        organization {
                        organizationID
                        name
                        }
                    }
                    }
                }
            """,
        )
        return result["updateService"]

    async def update_permissions_for_stream(self, service_id, stream_id, read, write):
        self._before_mutation()
        result = await self.conn.query_control(
            variables={
                "serviceID": service_id,
                "streamID": stream_id,
                "read": read,
                "write": write,
            },
            query="""
                mutation UpdateServicePermissions(
                    $serviceID: UUID!
                    $streamID: UUID!
                    $read: Boolean
                    $write: Boolean
                ) {
                    updateServiceStreamPermissions(
                        serviceID: $serviceID
                        streamID: $streamID
                        read: $read
                        write: $write
                    ) {
                        serviceID
                        streamID
                        read
                        write
                    }
                }
            """,
        )
        return result["updateServiceStreamPermissions"]

    async def delete(self, service_id):
        self._before_mutation()
        result = await self.conn.query_control(
            variables={
                "serviceID": service_id,
            },
            query="""
                mutation DeleteService($serviceID: UUID!) {
                    deleteService(serviceID: $serviceID)
                }
            """,
        )
        return result["deleteService"]

    async def issue_secret(self, service_id, description):
        self._before_mutation()
        result = await self.conn.query_control(
            variables={
                "serviceID": service_id,
                "description": description,
            },
            query="""
                mutation IssueServiceSecret($serviceID: UUID!, $description: String!) {
                    issueServiceSecret(serviceID: $serviceID, description: $description) {
                        token
                    }
                }
            """,
        )
        return result["issueServiceSecret"]

    async def list_secrets(self, service_id):
        result = await self.conn.query_control(
            variables={
                "serviceID": service_id,
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
            """,
        )
        return result["secretsForService"]
