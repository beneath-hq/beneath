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
        text text NOT NULL,
        bool boolean NOT NULL,
        timestamp timestamptz NOT NULL
    );
        """
    )


def insert_many_types(cursor, table):
    cursor.execute(f"INSERT INTO {table} VALUES (1, 'abc', True, now());")


def update_many_types(cursor, table):
    cursor.execute(f"UPDATE {table} SET text = 'def', timestamp = now() WHERE id = 1;")


def delete_many_types(cursor, table):
    cursor.execute(f"DELETE FROM {table} WHERE id = 1;")


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


# TODO: test tables with compound pks
def insert_compound_pk(cursor, table):
    ...


def update_compound_pk(cursor, table):
    ...


def delete_compound_pk(cursor, table):
    ...


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


# used for wal2json plugin
# TODO: update wal2json plugin, so I can include more parameters and get more data. Document the process.
def get_changes(cursor):
    cursor.execute(
        f"""
        SELECT data FROM pg_logical_slot_get_changes('{config['postgres']['replication_slot']}', NULL, NULL,
        'include-lsn', 'True', 'include-timestamp', 'True', 'add-tables', 'public.table1, public.table2');
        """
    )
    changes = cursor.fetchall()
    return changes


# used for pgoutput plugin
# TODO: figure out if I can decode the binary data into useful text data. Debezium somehow decodes it.
def get_binary_changes(cursor):
    cursor.execute(
        f"""
        SELECT data FROM pg_logical_slot_get_binary_changes('{config['postgres']['replication_slot'] + "_pgoutput"}', NULL, NULL,
        'proto_version', '1', 'publication_names', 'beneath_publication');
        """
    )
    changes = cursor.fetchall()
    return changes


def peek_binary_changes(cursor):
    cursor.execute(
        f"""
        SELECT data FROM pg_logical_slot_peek_binary_changes('{config['postgres']['replication_slot'] + "_pgoutput"}', NULL, NULL,
        'proto_version', '1', 'publication_names', 'beneath_publication');
        """
    )
    changes = cursor.fetchall()
    return changes


if __name__ == "__main__":
    # setup
    cursor = connect_to_source_db()
    drop_table(cursor, "table1")
    drop_table(cursor, "table2")
    create_table_simple(cursor, "table1")
    create_table_many_types(cursor, "table2")

    # operate
    insert_simple(cursor, "table1")
    insert_many_types(cursor, "table2")
    update_simple(cursor, "table1")
    update_many_types(cursor, "table2")
    delete_simple(cursor, "table1")
    # delete_many_types(cursor, "table2")

    # inspect
    # print(get_changes(cursor))
    # print(get_binary_changes(cursor))
    # print(peek_binary_changes(cursor))

    # cleanup
    # drop_table(cursor, "table1")
    # drop_table(cursor, "table2")
