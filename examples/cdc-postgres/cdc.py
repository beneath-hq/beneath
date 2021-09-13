# resources:
# https://debezium.io/documentation/reference/0.9/connectors/postgresql.html#how-the-postgresql-connector-works

import asyncio
import psycopg2
import yaml

# TODO: specify tables to replicate
with open(".development.yaml", "r") as ymlfile:
    config = yaml.safe_load(ymlfile)

POLLING_INTERVAL = 600  # 10 min


def connect_to_db():
    conn = psycopg2.connect(
        database="testdb",
        user=config["postgres"]["username"],
        password=config["postgres"]["password"],
        host=config["postgres"]["host"],
        port="5432",
    )
    conn.autocommit = True
    return conn


# TODO: implement.
# - create Beneath tables if they don't exist (need to infer schema)
# - write a mini transpilier to convert the DML commands (INSERT, UPDATE, DELETE) to Beneath writes
# - write a another transpiler to convert Postgres types to Beneath types
# - handle deletes and updates. idea: in the table schema, include a "delete" metacolumn, and set to true if the row is deleted or updated
# - mind schema changes (note that DDL changes / schema evolutions are not captured by logical decoding )
def write_changes_to_beneath(changes):
    print(changes)


# TODO: need to specify precise tables for replication
async def poll_replication_stream(connection):
    cursor = conn.cursor()

    while True:
        print("getting changes...")
        cursor.execute(
            "SELECT data FROM pg_logical_slot_get_changes('test_slot', NULL, NULL, 'pretty-print', '1');"
        )
        changes = cursor.fetchall()
        write_changes_to_beneath(changes)
        print("sleeping...")
        asyncio.sleep(POLLING_INTERVAL)


if __name__ == "__main__":
    conn = connect_to_db()

    asyncio.run(poll_replication_stream(conn))
