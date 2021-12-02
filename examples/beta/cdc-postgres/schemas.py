import json


def convert_pg_type_to_beneath_type(pg_type):
    if pg_type == "int32":
        # change this back to Int32 once I merge the bugfix into the stable branch
        # return "Int32"
        return "Int64"
    if pg_type == "string":
        return "String"
    if pg_type == "boolean":
        return "Boolean"
    if pg_type == "bytes":
        return "Bytes"
    raise Exception(f"Unrecognized type '{pg_type}' – need to implement")


def convert_pg_nullable_to_beneath_nullable(pg_nullable):
    if pg_nullable == False:
        return "!"
    return ""


def convert_pg_pk_to_beneath_pk(key_schema, pg_field):
    for key in key_schema:
        if key["field"] == pg_field:
            return " @key"
    return ""


def makeKeyForCheckpointedSchema(record):
    database = record["source_db"]
    namespace = record["source_schema"]
    table = record["source_table"]
    return f"{database}:{namespace}:{table}:schema"


async def get_schema(client, project_path, record):
    # get Postgres schema
    checkpoints = await client.checkpointer(
        project_path=project_path,
        metatable_name="checkpoints",
    )
    await client.start()
    dbz_schema_as_string = await checkpoints.get(makeKeyForCheckpointedSchema(record))

    # transpile Postgres schema langugage into Beneath schema langugage
    dbz_schema = json.loads(dbz_schema_as_string)
    key_schema = json.loads(dbz_schema["keySchema"])
    value_schema = json.loads(dbz_schema["valueSchema"])
    beneath_columns = []
    for pg_column in value_schema:
        beneath_column = f"{pg_column['field']}: {convert_pg_type_to_beneath_type(pg_column['type'])}{convert_pg_nullable_to_beneath_nullable(pg_column['optional'])}{convert_pg_pk_to_beneath_pk(key_schema, pg_column['field'])}"
        beneath_columns.append(beneath_column)

    # add metadata columns
    beneath_columns.append("_updated_at: Timestamp!")
    beneath_columns.append("_deleted_at: Timestamp")

    # build Beneath schema
    beneath_schema = "type {} @schema {{\n\t{}\n}}".format(
        record["source_table"], "\n\t".join(beneath_columns)
    )

    return beneath_schema
