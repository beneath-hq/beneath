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
    table: String! @key
    operation: String! @key
    timestamp: Timestamp! @key
    value: String!
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
            "SELECT data FROM pg_logical_slot_get_changes('test_slot', NULL, NULL);"
        )  # for production
        # cursor.execute(
        #     "SELECT data FROM pg_logical_slot_peek_changes('test_slot', NULL, NULL);"
        # )  # for testing
        data = cursor.fetchall()
        for txn in data:
            # TODO: what is txn[1] for? when does it get used?
            changes = json.loads(txn[0])["change"]
            for change in changes:
                value_dict = {}
                if change["kind"] == "insert":
                    value_dict["columnnames"] = change["columnnames"]
                    value_dict["columntypes"] = change["columntypes"]
                    value_dict["columnvalues"] = change["columnvalues"]
                elif change["kind"] == "update":
                    value_dict["columnnames"] = change["columnnames"]
                    value_dict["columntypes"] = change["columntypes"]
                    value_dict["columnvalues"] = change["columnvalues"]
                    value_dict["oldkeys"] = change["oldkeys"]
                elif change["kind"] == "delete":
                    value_dict["oldkeys"] = change["oldkeys"]
                yield {
                    "table": change["table"],
                    "operation": change["kind"],
                    "timestamp": datetime.now(),
                    "value": json.dumps(value_dict),
                }
        await asyncio.sleep(POLLING_INTERVAL)


def filter_for_table(table):
    async def filter(record):
        if record["table"] == table:
            yield record

    return filter


def fan_out(p, all_changes, list_of_tables):
    for table in list_of_tables:
        table_changes = p.apply(all_changes, filter_for_table(table))
        # TODO: figure out schemas
        p.write_table(
            table_changes,
            f"{config['beneath']['username']}/{config['beneath']['project']}/{table}-changes",
            schema=SCHEMA,
            description=f"Fan-out table ({table}) created by a Postgres CDC service",
        )


if __name__ == "__main__":
    p = beneath.Pipeline(parse_args=True, disable_checkpoints=True)
    p.description = "Postgres CDC"
    all_changes = p.generate(get_changes)
    # Q: should I persist these intermediate results?
    # reasons yes:
    # - whenever pg_logical_slot_get_changes() gets called, its cursor advances, so I better capture what it gave me
    # - if the fan-out fails (due to new table, schema evolution, types, etc), then I'll have captured the changes
    # reasons no:
    # - if my fan-out doesn't use table-specific schemas
    # - on restart, will I write full snapshots anyways?
    p.write_table(
        all_changes,
        f"{config['beneath']['username']}/{config['beneath']['project']}/all-changes",
        schema=SCHEMA,
        description="Table automatically created from a Postgres CDC service",
    )
    fan_out(p, all_changes, config["postgres"]["tables"])
    p.main()
