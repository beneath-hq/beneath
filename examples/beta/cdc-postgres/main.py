import asyncio
import beneath
import psycopg2
import json
import yaml
from datetime import datetime
from schemas import get_schema

with open(".development.yaml", "r") as ymlfile:
    config = yaml.safe_load(ymlfile)

POLLING_INTERVAL = 5
SCHEMA = """
type Table @schema {
    table: String! @key
    timestamp: Timestamp! @key
    operation: String! @key
    value: String!
}
"""


def connect_to_source_db():
    conn = psycopg2.connect(
        database=config["postgres"]["database"],
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
            f"""
        SELECT data FROM pg_logical_slot_get_changes('{config['postgres']['replication_slot']}', NULL, NULL,
        'include-lsn', 'True', 'include-timestamp', 'True', 'add-tables', '{','.join(config['postgres']['tables'])}');
        """
        )
        txns = cursor.fetchall()
        for txn in txns:
            txn_json = json.loads(txn[0])
            changes = txn_json["change"]
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
                    "timestamp": datetime.strptime(
                        f"{txn_json['timestamp']}00", "%Y-%m-%d %H:%M:%S.%f%z"
                    ),
                    "operation": change["kind"],
                    "value": json.dumps(value_dict),
                }
        # TODO: consider checkpointing the LSN (but pg_logical_slot_get_changes() doesn't let us choose LSN position)
        # p.checkpoints.set("nextlsn", txn_json["nextlsn"])
        await asyncio.sleep(POLLING_INTERVAL)


def filter_for_table(table):
    async def filter(record):
        if record["table"] == table:
            # TODO: convert general CDC record to table-specific schema
            yield record

    return filter


def fan_out(p, all_changes, list_of_tables):
    # list_of_tables: ["schemaA.table1", "schemaA.table2", "schemaB.table1", ...]
    for schema_table in list_of_tables:
        table = schema_table.split(".")[1]
        table_changes = p.apply(all_changes, filter_for_table(table))
        p.write_table(
            table_changes,
            f"{config['beneath']['username']}/{config['beneath']['project']}/{table}-changes",
            schema=get_schema(cursor, table),
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
