# resources:
# https://debezium.io/documentation/reference/0.9/connectors/postgresql.html#how-the-postgresql-connector-works
# https://dba.stackexchange.com/questions/219397/postgres-logical-replication-for-specific-tables # no answer

import asyncio
import beneath
import psycopg2
import json
import yaml
from datetime import datetime

with open(".development.yaml", "r") as ymlfile:
    config = yaml.safe_load(ymlfile)

POLLING_INTERVAL = 30
SCHEMA = """
type Table @schema {
    timestamp: Timestamp! @key
    operation: String!
    table: String!
    columnnames: String!
    columntypes: String!
    columnvalues: String!
}
"""


def connect_to_source_db():
    conn = psycopg2.connect(
        database="testdb",
        user=config["postgres"]["username"],
        password=config["postgres"]["password"],
        host=config["postgres"]["host"],
        port="5432",
    )
    conn.autocommit = True
    return conn


conn = connect_to_source_db()
cursor = conn.cursor()


async def get_changes(p):
    while True:
        cursor.execute(
            "SELECT data FROM pg_logical_slot_get_changes('test_slot', NULL, NULL);"  # for production
            # "SELECT data FROM pg_logical_slot_peek_changes('test_slot', NULL, NULL);"  # for testing
        )
        changes = cursor.fetchall()
        for change in changes:
            data = json.loads(change[0])["change"][0]
            # TODO: figure out better schema and fan out
            yield {
                "timestamp": datetime.now(),
                "operation": data["kind"],
                "table": data["table"],
                "columnnames": str(data["columnnames"]),
                "columntypes": str(data["columntypes"]),
                "columnvalues": str(data["columnvalues"]),
            }
        await asyncio.sleep(POLLING_INTERVAL)


if __name__ == "__main__":
    p = beneath.Pipeline(parse_args=True, disable_checkpoints=True)
    p.description = "Postgres CDC"
    changes = p.generate(get_changes)
    p.write_table(
        changes,
        f"{config['beneath']['username']}/{config['beneath']['project']}/root",
        schema=SCHEMA,
        description="Table automatically created from a Postgres CDC service",
    )
    p.main()
