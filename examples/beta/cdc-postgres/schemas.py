# Q: should I somehow leverage code in the Beneath Python client for inferring schema?
def convert_pg_type_to_beneath_type(pg_type):
    if pg_type == "integer":
        return "Int32"
    if pg_type == "timestamp with time zone":
        return "Timestamp"
    if pg_type == "boolean":
        return "Boolean"
    if pg_type == "text":
        return "String"
    raise Exception(f"Unrecognized type '{pg_type}' – need to implement")


def convert_pg_nullable_to_beneath_nullable(pg_nullable):
    if pg_nullable == "NO":
        return "!"
    return ""


def convert_pg_pk_to_beneath_pk(pg_pk):
    if pg_pk:
        return " @key"
    return ""


def get_schema(cursor, table):
    # get Postgres schema
    cursor.execute(
        f"""
      SELECT c.column_name, c.data_type, is_nullable,
      CASE WHEN EXISTS(SELECT 1 FROM INFORMATION_SCHEMA.constraint_column_usage k WHERE c.table_name = k.table_name and k.column_name = c.column_name)
          THEN true ELSE false END as primary_key
      FROM INFORMATION_SCHEMA.COLUMNS c 
      WHERE c.table_name='{table}';
      """
    )
    pg_schema = cursor.fetchall()

    # transpile Postgres schema langugage into Beneath schema langugage
    beneath_columns = []
    for pg_column in pg_schema:
        beneath_column = f"{pg_column[0]}: {convert_pg_type_to_beneath_type(pg_column[1])}{convert_pg_nullable_to_beneath_nullable(pg_column[2])}{convert_pg_pk_to_beneath_pk(pg_column[3])}"
        beneath_columns.append(beneath_column)

    # build Beneath schema
    beneath_schema = "type {} @schema {{\n\t{}\n}}".format(
        table, "\n\t".join(beneath_columns)
    )
    return beneath_schema
