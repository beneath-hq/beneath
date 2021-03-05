from beneath.admin.base import _ResourceBase
from beneath.utils import format_entity_name


class Streams(_ResourceBase):
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
                        meta
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
        self,
        organization_name,
        project_name,
        stream_name,
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
                        meta
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

    async def find_instance_by_organization_project_name_and_version(
        self,
        organization_name,
        project_name,
        stream_name,
        version,
    ):
        result = await self.conn.query_control(
            variables={
                "organizationName": format_entity_name(organization_name),
                "projectName": format_entity_name(project_name),
                "streamName": format_entity_name(stream_name),
                "version": version,
            },
            query="""
                query StreamInstanceByOrganizationProjectStreamAndVersion(
                    $organizationName: String!
                    $projectName: String!
                    $streamName: String!
                    $version: Int!
                ) {
                    streamInstanceByOrganizationProjectStreamAndVersion(
                        organizationName: $organizationName
                        projectName: $projectName
                        streamName: $streamName
                        version: $version
                    ) {
                        streamInstanceID
                        stream {
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
                            meta
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
                        streamID
                        version
                        createdOn
                        madePrimaryOn
                        madeFinalOn
                    }
                }
            """,
        )
        return result["streamInstanceByOrganizationProjectStreamAndVersion"]

    async def find_instance(self, stream_id, version):
        result = await self.conn.query_control(
            variables={
                "streamID": stream_id,
                "version": version,
            },
            query="""
                query StreamInstanceForStream($streamID: UUID!) {
                    streamInstanceForStream(streamID: $streamID) {
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

    async def compile_schema(self, schema_kind, schema, indexes=None):
        result = await self.conn.query_control(
            variables={
                "input": {
                    "schemaKind": schema_kind,
                    "schema": schema,
                    "indexes": indexes,
                },
            },
            query="""
                query CompileSchema($input: CompileSchemaInput!) {
                    compileSchema(input: $input) {
                        canonicalAvroSchema
                        canonicalIndexes
                    }
                }
            """,
        )
        return result["compileSchema"]

    async def create(
        self,
        organization_name,
        project_name,
        stream_name,
        schema_kind,
        schema,
        indexes=None,
        description=None,
        meta=None,
        allow_manual_writes=None,
        use_log=None,
        use_index=None,
        use_warehouse=None,
        log_retention_seconds=None,
        index_retention_seconds=None,
        warehouse_retention_seconds=None,
        update_if_exists=None,
    ):
        self._before_mutation()
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
                    "meta": meta,
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
                        meta
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
        self._before_mutation()
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
        version: int = None,
        make_primary=None,
        update_if_exists=None,
    ):
        self._before_mutation()
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
        self._before_mutation()
        result = await self.conn.query_control(
            variables={
                "input": {
                    "streamInstanceID": instance_id,
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
        self._before_mutation()
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
