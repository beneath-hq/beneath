import asyncio
import random
import string
import os
import sys
import inspect
import yaml

# add parent directory to path, so I can import from the cdc.py module
currentdir = os.path.dirname(os.path.abspath(inspect.getfile(inspect.currentframe())))
parentdir = os.path.dirname(currentdir)
sys.path.insert(0, parentdir)

from cdc import connect_to_source_db

with open(".development.yaml", "r") as ymlfile:
    config = yaml.safe_load(ymlfile)

SLEEP_INTERVAL = 10  # seconds
OUTPUT_PLUGIN = "pgoutput"  # or "wal2json"


def create_table(cursor):
    cursor.execute(
        """
    CREATE TABLE users (
        user_id int,
        email text,
        authorized boolean,
        created_on timestamptz, 
        updated_on timestamptz
    );
    """
    )


def drop_table(cursor):
    cursor.execute("DROP TABLE IF EXISTS users;")


def do_everything_txn(cursor):
    cursor.execute(
        f"""
    DROP TABLE IF EXISTS actions;
    CREATE TABLE actions (a integer primary key);
    INSERT INTO actions (a) VALUES(1);
    UPDATE actions SET a = 2 WHERE a = 1;
    DELETE FROM actions WHERE a = 2;
    TRUNCATE TABLE actions;
    """
    )


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


def pick_random_op(cursor):
    random.choice([insert, update, delete])(cursor)


async def do_things_continuously(cursor):
    while True:
        pick_random_op(cursor)
        await asyncio.sleep(SLEEP_INTERVAL)


def get_changes(cursor):
    cursor.execute(
        f"SELECT data FROM pg_logical_slot_get_changes('{config['postgres']['replication_slot']}', NULL, NULL);"
    )
    changes = cursor.fetchall()
    return changes


if __name__ == "__main__":
    # setup
    conn = connect_to_source_db()
    cursor = conn.cursor()
    create_table(cursor)

    # operate
    do_everything_txn(cursor)
    # asyncio.run(do_things_continuously(cursor))

    # inspect
    # print(get_changes(cursor))

    # cleanup
    drop_table(cursor)
