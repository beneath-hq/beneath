from beneath.connection import Connection
from beneath.utils import format_entity_name


class Streams:
    def __init__(self, conn: Connection):
        self.conn = conn

    async def find_by_id(self, stream_id):
        result = await self.conn.query_control(
            variables={
                "streamID": stream_id,
            },
            query="""
                query StreamByID($streamID: UUID!) {
                    streamByID(streamID: $streamID) {
                        streamID
                        name
                        description
                        createdOn
                        updatedOn
                        project {
                            projectID
                            name
                        }
                        schemaKind
                        schema
                        avroSchema
                        streamIndexes {
                            fields
                            primary
                            normalize
                        }
                        allowManualWrites
                        useLog
                        useIndex
                        useWarehouse
                        logRetentionSeconds
                        indexRetentionSeconds
                        warehouseRetentionSeconds
                        primaryStreamInstanceID
                        primaryStreamInstance {
                            streamInstanceID
                            createdOn
                            version
                            madePrimaryOn
                            madeFinalOn
                        }
                        instancesCreatedCount
                        instancesDeletedCount
                        instancesMadeFinalCount
                        instancesMadePrimaryCount
                    }
                }
            """,
        )
        return result["streamByID"]

    async def find_by_organization_project_and_name(
        self, organization_name, project_name, stream_name
    ):
        result = await self.conn.query_control(
            variables={
                "organizationName": format_entity_name(organization_name),
                "projectName": format_entity_name(project_name),
                "streamName": format_entity_name(stream_name),
            },
            query="""
                query StreamByOrganizationProjectAndName(
                    $organizationName: String!
                    $projectName: String!
                    $streamName: String!
                ) {
                    streamByOrganizationProjectAndName(
                        organizationName: $organizationName
                        projectName: $projectName
                        streamName: $streamName
                    ) {
                        streamID
                        name
                        description
                        createdOn
                        updatedOn
                        project {
                            projectID
                            name
                        }
                        schemaKind
                        schema
                        avroSchema
                        streamIndexes {
                            fields
                            primary
                            normalize
                        }
                        allowManualWrites
                        useLog
                        useIndex
                        useWarehouse
                        logRetentionSeconds
                        indexRetentionSeconds
                        warehouseRetentionSeconds
                        primaryStreamInstance {
                            streamInstanceID
                            createdOn
                            version
                            madePrimaryOn
                            madeFinalOn
                        }
                        instancesCreatedCount
                        instancesDeletedCount
                        instancesMadeFinalCount
                        instancesMadePrimaryCount
                    }
                }
            """,
        )
        return result["streamByOrganizationProjectAndName"]

    async def find_instances(self, stream_id):
        result = await self.conn.query_control(
            variables={
                "streamID": stream_id,
            },
            query="""
                query StreamInstancesForStream($streamID: UUID!) {
                    streamInstancesForStream(streamID: $streamID) {
                        streamInstanceID
                        streamID
                        createdOn
                        version
                        madePrimaryOn
                        madeFinalOn
                    }
                }
            """,
        )
        return result["streamInstancesForStream"]

    async def create(
        self,
        organization_name,
        project_name,
        stream_name,
        schema_kind,
        schema,
        indexes=None,
        description=None,
        allow_manual_writes=None,
        use_log=None,
        use_index=None,
        use_warehouse=None,
        log_retention_seconds=None,
        index_retention_seconds=None,
        warehouse_retention_seconds=None,
        update_if_exists=None,
    ):
        result = await self.conn.query_control(
            variables={
                "input": {
                    "organizationName": format_entity_name(organization_name),
                    "projectName": format_entity_name(project_name),
                    "streamName": format_entity_name(stream_name),
                    "schemaKind": schema_kind,
                    "schema": schema,
                    "indexes": indexes,
                    "description": description,
                    "allowManualWrites": allow_manual_writes,
                    "useLog": use_log,
                    "useIndex": use_index,
                    "useWarehouse": use_warehouse,
                    "logRetentionSeconds": log_retention_seconds,
                    "indexRetentionSeconds": index_retention_seconds,
                    "warehouseRetentionSeconds": warehouse_retention_seconds,
                    "updateIfExists": update_if_exists,
                },
            },
            query="""
                mutation CreateStream($input: CreateStreamInput!) {
                    createStream(input: $input) {
                        streamID
                        name
                        description
                        createdOn
                        updatedOn
                        project {
                            projectID
                            name
                        }
                        schemaKind
                        schema
                        avroSchema
                        streamIndexes {
                            fields
                            primary
                            normalize
                        }
                        allowManualWrites
                        useLog
                        useIndex
                        useWarehouse
                        logRetentionSeconds
                        indexRetentionSeconds
                        warehouseRetentionSeconds
                        primaryStreamInstanceID
                        primaryStreamInstance {
                            streamInstanceID
                            createdOn
                            version
                            madePrimaryOn
                            madeFinalOn
                        }
                        instancesCreatedCount
                        instancesDeletedCount
                        instancesMadeFinalCount
                        instancesMadePrimaryCount
                    }
                }
            """,
        )
        return result["createStream"]

    async def delete(self, stream_id):
        result = await self.conn.query_control(
            variables={
                "streamID": stream_id,
            },
            query="""
                mutation DeleteStream($streamID: UUID!) {
                    deleteStream(streamID: $streamID)
                }
            """,
        )
        return result["deleteStream"]

    async def create_instance(
        self,
        stream_id,
        version: int,
        make_primary=None,
        update_if_exists=None,
    ):
        result = await self.conn.query_control(
            variables={
                "input": {
                    "streamID": stream_id,
                    "version": version,
                    "makePrimary": make_primary,
                    "updateIfExists": update_if_exists,
                },
            },
            query="""
                mutation CreateStreamInstance($input: CreateStreamInstanceInput!) {
                    createStreamInstance(input: $input) {
                        streamInstanceID
                        streamID
                        createdOn
                        version
                        madePrimaryOn
                        madeFinalOn
                    }
                }
            """,
        )
        return result["createStreamInstance"]

    async def update_instance(self, instance_id, make_final=None, make_primary=None):
        result = await self.conn.query_control(
            variables={
                "input": {
                    "instanceID": instance_id,
                    "makeFinal": make_final,
                    "makePrimary": make_primary,
                },
            },
            query="""
                mutation UpdateStreamInstance($input: UpdateStreamInstanceInput!) {
                    updateStreamInstance(input: $input) {
                        streamInstanceID
                        streamID
                        createdOn
                        version
                        madePrimaryOn
                        madeFinalOn
                    }
                }
            """,
        )
        return result["updateStreamInstance"]

    async def delete_instance(self, instance_id):
        result = await self.conn.query_control(
            variables={
                "instanceID": instance_id,
            },
            query="""
                mutation deleteStreamInstance($instanceID: UUID!) {
                    deleteStreamInstance(instanceID: $instanceID)
                }
            """,
        )
        return result["deleteStreamInstance"]
