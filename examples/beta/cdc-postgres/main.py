import asyncio
import beneath
import psycopg2
import json
import yaml
from datetime import datetime
from schemas import get_schema, check_for_and_encode_ts

with open(".development.yaml", "r") as ymlfile:
    config = yaml.safe_load(ymlfile)

POLLING_INTERVAL = 5
SCHEMA = """
type Change @schema {
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
    cursor = conn.cursor()
    return cursor


cursor = connect_to_source_db()


async def get_all_changes(p):
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
                yield {
                    "table": change["table"],
                    "timestamp": datetime.strptime(
                        f"{txn_json['timestamp']}00", "%Y-%m-%d %H:%M:%S.%f%z"
                    ),
                    "operation": change["kind"],
                    "value": json.dumps(
                        {
                            k: change.get(k)
                            for k in (
                                "columnnames",
                                "columntypes",
                                "columnvalues",
                                "oldkeys",
                            )
                        }
                    ),
                }
        # TODO: consider checkpointing the LSN (but pg_logical_slot_get_changes() doesn't let us choose LSN position)
        # p.checkpoints.set("nextlsn", txn_json["nextlsn"])
        await asyncio.sleep(POLLING_INTERVAL)


def filter_for_table(table):
    async def filter(in_record):
        if in_record["table"] == table:

            # construct out_record
            value_blob = json.loads(in_record["value"])
            if in_record["operation"] in ["insert", "update"]:
                # TODO: test more types, might have to do more type conversions
                out_record = {
                    col: check_for_and_encode_ts(
                        value_blob["columntypes"][i], value_blob["columnvalues"][i]
                    )
                    for (i, col) in enumerate(value_blob["columnnames"])
                }
                out_record["_updated_at"] = in_record["timestamp"]
            if in_record["operation"] == "delete":
                # TODO: Handle required non-key columns.
                # - Option1: Set Replica Identity to FULL for all tables
                # -- not ideal, since a) requires more user setup and b) passes more data
                # - Option2: Just generate synthetic data here
                # -- the row is getting deleted anyways, so doesn't really matter what the values are
                # print(value_blob)
                out_record = dict(
                    zip(
                        value_blob["oldkeys"]["keynames"],
                        value_blob["oldkeys"]["keyvalues"],
                    )
                )
                out_record["_updated_at"] = in_record["timestamp"]
                out_record["_deleted_at"] = in_record["timestamp"]

            yield out_record

    return filter


def fan_out(p, all_changes, list_of_tables):
    # list_of_tables: ["schemaA.table1", "schemaA.table2", "schemaB.table1", ...]
    for schema_table in list_of_tables:
        schema = schema_table.split(".")[0]  # a Postgres "schema" (a namespace)
        table = schema_table.split(".")[1]
        table_changes = p.apply(all_changes, filter_for_table(table))
        p.write_table(
            table_changes,
            f"{config['beneath']['username']}/{config['beneath']['project']}/{config['postgres']['database']}-{schema}-{table}",
            schema=get_schema(cursor, table),  # a Beneath "schema" (type info)
            description=f"{table} table replicated from Postgres",
        )


if __name__ == "__main__":
    p = beneath.Pipeline(parse_args=True, disable_checkpoints=True)
    p.description = "Postgres CDC"
    all_changes = p.generate(get_all_changes)
    p.write_table(
        all_changes,
        f"{config['beneath']['username']}/{config['beneath']['project']}/{config['postgres']['database']}-cdc",
        schema=SCHEMA,
        description="Raw data captured from a Postgres CDC service",
    )
    fan_out(p, all_changes, config["postgres"]["tables"])
    p.main()
