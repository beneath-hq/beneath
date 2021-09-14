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

from cdc import connect_to_source_db

SLEEP_INTERVAL = 10  # seconds


def insert(cursor):
    user_id = random.randint(0, 99999)
    bool = random.choice(["True", "False"])
    username = random.choice(string.ascii_letters)
    cursor.execute(
        f"""
    INSERT INTO users VALUES ({user_id}, '{username}@test.com', {bool}, now(), now());
        """
    )
    print(f"Inserted dummy record (id: {user_id})")


def update(cursor):
    cursor.execute("UPDATE users SET updated_on = now() WHERE authorized = True;")
    print("Updated authorized users")


def delete(cursor):
    username = random.choice(string.ascii_letters)
    cursor.execute(
        f"""
    DELETE FROM users WHERE email like '{username}%';
    """
    )
    print(f"Deleted users where email starts with {username}")


def pick_op(cursor):
    random.choice([insert, update, delete])(cursor)


async def do_things(conn):
    cursor = conn.cursor()
    while True:
        pick_op(cursor)
        await asyncio.sleep(SLEEP_INTERVAL)


if __name__ == "__main__":
    conn = connect_to_source_db()
    asyncio.run(do_things(conn))