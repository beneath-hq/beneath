import asyncio
import beneath


async def main():
    client = beneath.Client()
    await client.start()

    table = await client.find_table("examples/reddit/r-wallstreetbets-comments-scores")
    instance = await table.find_instance(version=1)  # primary_version + 1

    # I'm deduplicating month-by-month because otherwise I hit the max_bytes read limit when I read the results from the warehouse query
    # Currently, the query_warehouse config only controls for how much data is scanned, not how much data is read from the results
    for month_number in ["03", "04", "05", "06", "07", "08", "09"]:
        print(f"running query for month {month_number}...")
        deduplicated = await beneath.query_warehouse(
            f"""
        SELECT r.* FROM (
          SELECT array_agg(t ORDER BY t.__timestamp DESC LIMIT 1)[OFFSET(0)] r
            FROM (
                select s.id, s.score, s.__timestamp, s.__key
                from `examples/reddit/r-wallstreetbets-comments-scores` s
                left join `examples/reddit/r-wallstreetbets-comments` c on s.id = c.id
                where timestamp_trunc(c.created_on, month) = "2021-{month_number}-01"
            ) t
          GROUP BY t.__key
        )
      """
        )
        await instance.write(deduplicated.to_dict("records"))

    await client.stop()


asyncio.run(main())