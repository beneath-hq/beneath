import os
import json
import asyncio
import beneath
from postgres.schemas import get_schema

DATABASE_DBNAME = os.getenv("DATABASE_DBNAME")


async def main():
    client = beneath.Client()
    await client.start()

    # get project path
    me = await client.admin.organizations.find_me()
    PROJECT_PATH = f"{me['name']}/cdc-postgres-{DATABASE_DBNAME}"

    consumer = await client.consumer(
        f"{PROJECT_PATH}/raw-changes",
        subscription_path=f"{PROJECT_PATH}/fanout-service",
    )

    async def fanout(in_record):
        # r = read (from a snapshot), c = create, u = update
        if in_record["op"] in ["r", "c", "u"]:
            out_record = json.loads(in_record["after"])
            out_record["_updated_at"] = in_record["source_ts_ms"]
        # d = delete
        if in_record["op"] == "d":
            out_record = json.loads(in_record["before"])
            out_record["_updated_at"] = in_record["source_ts_ms"]
            out_record["_deleted_at"] = in_record["source_ts_ms"]

        table = await client.create_table(
            table_path=f"{PROJECT_PATH}/{in_record['source_schema']}-{in_record['source_table']}",
            schema=await get_schema(client, PROJECT_PATH, in_record),
            update_if_exists=True,
            description=f"Replicated from Postgres.",
        )
        instance = table.primary_instance
        await instance.write(out_record)

    await consumer.subscribe(cb=fanout)


if __name__ == "__main__":
    asyncio.run(main())
