import inspect
import sys
import os
import yaml

# add parent directory to path, so I can import from the main.py module
currentdir = os.path.dirname(os.path.abspath(inspect.getfile(inspect.currentframe())))
parentdir = os.path.dirname(currentdir)
sys.path.insert(0, parentdir)

from main import connect_to_source_db

with open(".development.yaml", "r") as ymlfile:
    config = yaml.safe_load(ymlfile)

SLEEP_INTERVAL = 10  # seconds
OUTPUT_PLUGIN = "pgoutput"  # or "wal2json"


def create_table_simple(cursor, table):
    cursor.execute(
        f"""
    CREATE TABLE {table} (
        id int primary key,
        val text
    );
    """
    )


def insert_simple(cursor, table):
    cursor.execute(f"INSERT INTO {table} VALUES (1, 'abc');")


def update_simple(cursor, table):
    cursor.execute(f"UPDATE {table} SET val = 'def' WHERE id = 1;")


def delete_simple(cursor, table):
    cursor.execute(f"DELETE FROM {table} WHERE id = 1;")


def create_table_many_types(cursor, table):
    cursor.execute(
        f"""
    CREATE TABLE {table} (
        id int primary key,
        text text,
        bool boolean,
        timestamp timestamptz
    );
        """
    )


def insert_many_types(cursor, table):
    cursor.execute(f"INSERT INTO {table} VALUES (1, 'abc', True, now());")


def create_table_compound_pk(cursor, table):
    cursor.execute(
        f"""
    CREATE TABLE {table} (
        id_1 int,
        id_2 int,
        val text,
        primary key(id_1, id_2)
    );
    """
    )


def drop_table(cursor, table):
    cursor.execute(f"DROP TABLE IF EXISTS {table};")


def do_everything_txn(cursor, table):
    cursor.execute(
        f"""
    DROP TABLE IF EXISTS {table};
    CREATE TABLE {table} (a integer primary key);
    INSERT INTO {table} (a) VALUES(1);
    UPDATE {table} SET a = 2 WHERE a = 1;
    DELETE FROM {table} WHERE a = 2;
    TRUNCATE TABLE {table};
    """
    )


def get_changes(cursor):
    # include-pk, 'True'
    cursor.execute(
        f"""
        SELECT data FROM pg_logical_slot_get_changes('{config['postgres']['replication_slot']}', NULL, NULL,
        'include-lsn', 'True', 'include-timestamp', 'True', 'add-tables', 'public.table1, public.table2');
        """
    )
    changes = cursor.fetchall()
    return changes


if __name__ == "__main__":
    # setup
    conn = connect_to_source_db()
    cursor = conn.cursor()
    drop_table(cursor, "table1")
    drop_table(cursor, "table2")
    create_table_simple(cursor, "table1")
    create_table_simple(cursor, "table2")

    # operate
    insert_simple(cursor, "table1")
    update_simple(cursor, "table1")
    delete_simple(cursor, "table1")

    # inspect
    # print(get_changes(cursor))

    # cleanup
    # drop_table(cursor, "table1")
    # drop_table(cursor, "table2")
