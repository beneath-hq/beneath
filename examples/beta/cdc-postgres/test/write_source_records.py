import asyncio
import random
import string
import os
import sys
import inspect

# add parent directory to path, so I can import from the cdc.py module
currentdir = os.path.dirname(os.path.abspath(inspect.getfile(inspect.currentframe())))
parentdir = os.path.dirname(currentdir)
sys.path.insert(0, parentdir)

from cdc import connect_to_db

SLEEP_INTERVAL = 5  # seconds

# TODO: test updates and deletes, too
async def stream_records(conn):
    cursor = conn.cursor()

    count = 0
    while True:
        bool = random.choice(["True", "False"])
        username = random.choice(string.ascii_letters)
        cursor.execute(
            f"""
        INSERT INTO users VALUES ({count}, '{username}@test.com', {bool}, now(), now());
          """
        )
        count += 1
        print(f"Inserted dummy record (#{count})")
        await asyncio.sleep(SLEEP_INTERVAL)


if __name__ == "__main__":
    conn = connect_to_db()
    asyncio.run(stream_records(conn))