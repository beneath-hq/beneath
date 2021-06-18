from beneath.admin.base import _ResourceBase
from beneath.utils import format_entity_name


class Tables(_ResourceBase):
    async def find_by_id(self, table_id):
        result = await self.conn.query_control(
            variables={
                "tableID": table_id,
            },
            query="""
                query TableByID($tableID: UUID!) {
                    tableByID(tableID: $tableID) {
                        tableID
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
                        tableIndexes {
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
                        primaryTableInstanceID
                        primaryTableInstance {
                            tableInstanceID
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
        return result["tableByID"]

    async def find_by_organization_project_and_name(
        self,
        organization_name,
        project_name,
        table_name,
    ):
        result = await self.conn.query_control(
            variables={
                "organizationName": format_entity_name(organization_name),
                "projectName": format_entity_name(project_name),
                "tableName": format_entity_name(table_name),
            },
            query="""
                query TableByOrganizationProjectAndName(
                    $organizationName: String!
                    $projectName: String!
                    $tableName: String!
                ) {
                    tableByOrganizationProjectAndName(
                        organizationName: $organizationName
                        projectName: $projectName
                        tableName: $tableName
                    ) {
                        tableID
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
                        tableIndexes {
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
                        primaryTableInstance {
                            tableInstanceID
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
        return result["tableByOrganizationProjectAndName"]

    async def find_instance_by_organization_project_name_and_version(
        self,
        organization_name,
        project_name,
        table_name,
        version,
    ):
        result = await self.conn.query_control(
            variables={
                "organizationName": format_entity_name(organization_name),
                "projectName": format_entity_name(project_name),
                "tableName": format_entity_name(table_name),
                "version": version,
            },
            query="""
                query TableInstanceByOrganizationProjectTableAndVersion(
                    $organizationName: String!
                    $projectName: String!
                    $tableName: String!
                    $version: Int!
                ) {
                    tableInstanceByOrganizationProjectTableAndVersion(
                        organizationName: $organizationName
                        projectName: $projectName
                        tableName: $tableName
                        version: $version
                    ) {
                        tableInstanceID
                        table {
                            tableID
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
                            tableIndexes {
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
                            primaryTableInstance {
                                tableInstanceID
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
                        tableID
                        version
                        createdOn
                        madePrimaryOn
                        madeFinalOn
                    }
                }
            """,
        )
        return result["tableInstanceByOrganizationProjectTableAndVersion"]

    async def find_instance(self, table_id, version):
        result = await self.conn.query_control(
            variables={
                "tableID": table_id,
                "version": version,
            },
            query="""
                query TableInstanceForTable($tableID: UUID!) {
                    tableInstanceForTable(tableID: $tableID) {
                        tableInstanceID
                        tableID
                        createdOn
                        version
                        madePrimaryOn
                        madeFinalOn
                    }
                }
            """,
        )
        return result["tableInstancesForTable"]

    async def find_instances(self, table_id):
        result = await self.conn.query_control(
            variables={
                "tableID": table_id,
            },
            query="""
                query TableInstancesForTable($tableID: UUID!) {
                    tableInstancesForTable(tableID: $tableID) {
                        tableInstanceID
                        tableID
                        createdOn
                        version
                        madePrimaryOn
                        madeFinalOn
                    }
                }
            """,
        )
        return result["tableInstancesForTable"]

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
        table_name,
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
                    "tableName": format_entity_name(table_name),
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
                mutation CreateTable($input: CreateTableInput!) {
                    createTable(input: $input) {
                        tableID
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
                        tableIndexes {
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
                        primaryTableInstanceID
                        primaryTableInstance {
                            tableInstanceID
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
        return result["createTable"]

    async def delete(self, table_id):
        self._before_mutation()
        result = await self.conn.query_control(
            variables={
                "tableID": table_id,
            },
            query="""
                mutation DeleteTable($tableID: UUID!) {
                    deleteTable(tableID: $tableID)
                }
            """,
        )
        return result["deleteTable"]

    async def create_instance(
        self,
        table_id,
        version: int = None,
        make_primary=None,
        update_if_exists=None,
    ):
        self._before_mutation()
        result = await self.conn.query_control(
            variables={
                "input": {
                    "tableID": table_id,
                    "version": version,
                    "makePrimary": make_primary,
                    "updateIfExists": update_if_exists,
                },
            },
            query="""
                mutation CreateTableInstance($input: CreateTableInstanceInput!) {
                    createTableInstance(input: $input) {
                        tableInstanceID
                        tableID
                        createdOn
                        version
                        madePrimaryOn
                        madeFinalOn
                    }
                }
            """,
        )
        return result["createTableInstance"]

    async def update_instance(self, instance_id, make_final=None, make_primary=None):
        self._before_mutation()
        result = await self.conn.query_control(
            variables={
                "input": {
                    "tableInstanceID": instance_id,
                    "makeFinal": make_final,
                    "makePrimary": make_primary,
                },
            },
            query="""
                mutation UpdateTableInstance($input: UpdateTableInstanceInput!) {
                    updateTableInstance(input: $input) {
                        tableInstanceID
                        tableID
                        createdOn
                        version
                        madePrimaryOn
                        madeFinalOn
                    }
                }
            """,
        )
        return result["updateTableInstance"]

    async def delete_instance(self, instance_id):
        self._before_mutation()
        result = await self.conn.query_control(
            variables={
                "instanceID": instance_id,
            },
            query="""
                mutation deleteTableInstance($instanceID: UUID!) {
                    deleteTableInstance(instanceID: $instanceID)
                }
            """,
        )
        return result["deleteTableInstance"]
