import psycopg2
import yaml

with open(".development.yaml", "r") as ymlfile:
    config = yaml.safe_load(ymlfile)


def connect_to_source_db():
    conn = psycopg2.connect(
        database=config["postgres"]["database"],
        user=config["postgres"]["username"],
        password=config["postgres"]["password"],
        host=config["postgres"]["host"],
        port="5432",
    )
    conn.autocommit = True
    cur = conn.cursor()
    return cur


def create_table_simple(cur, table):
    cur.execute(
        f"""
    CREATE TABLE {table} (
        id int primary key,
        val text
    );
    """
    )


def insert_simple(cur, table):
    cur.execute(f"INSERT INTO {table} VALUES (1, 'abc');")


def update_simple(cur, table):
    cur.execute(f"UPDATE {table} SET val = 'def' WHERE id = 1;")


def delete_simple(cur, table):
    cur.execute(f"DELETE FROM {table} WHERE id = 1;")


def create_table_many_types(cur, table):
    cur.execute(
        f"""
    CREATE TABLE {table} (
        uuid uuid primary key,
        number int NOT NULL,
        text text NOT NULL,
        bool boolean NOT NULL,
        timestamp timestamptz NOT NULL
    );
        """
    )


def insert_many_types(cur, table):
    cur.execute(
        f"INSERT INTO {table} VALUES (uuid_generate_v4(), 1, 'abc', True, now());"
    )


def update_many_types(cur, table):
    cur.execute(f"UPDATE {table} SET text = 'def', timestamp = now() WHERE number = 1;")


def delete_many_types(cur, table):
    cur.execute(f"DELETE FROM {table} WHERE number = 1;")


def create_table_compound_pk(cur, table):
    cur.execute(
        f"""
    CREATE TABLE {table} (
        id_1 int,
        id_2 int,
        val text,
        primary key(id_1, id_2)
    );
    """
    )


def insert_compound_pk(cur, table):
    cur.execute(f"INSERT INTO {table} VALUES (1, 1, 'abc');")


def update_compound_pk(cur, table):
    cur.execute(f"UPDATE {table} SET val = 'def' WHERE id_1 = 1 AND id_2 = 1;")


def delete_compound_pk(cur, table):
    cur.execute(f"DELETE FROM {table} WHERE id_1 = 1 AND id_2 = 1;")


def do_everything_txn(cur, table):
    cur.execute(
        f"""
    CREATE TABLE {table} (a integer primary key);
    INSERT INTO {table} (a) VALUES(1);
    UPDATE {table} SET a = 2 WHERE a = 1;
    DELETE FROM {table} WHERE a = 2;
    TRUNCATE TABLE {table};
    """
    )


def drop_table(cur, table):
    cur.execute(f"DROP TABLE IF EXISTS {table};")


if __name__ == "__main__":
    # setup
    cur = connect_to_source_db()
    create_table_simple(cur, "table1")
    create_table_many_types(cur, "table2")

    # operate
    insert_simple(cur, "table1")
    update_simple(cur, "table1")
    delete_simple(cur, "table1")
    insert_many_types(cur, "table2")
    update_many_types(cur, "table2")
    delete_many_types(cur, "table2")
    do_everything_txn(cur, "table3")

    # cleanup
    drop_table(cur, "table1")
    drop_table(cur, "table2")
    drop_table(cur, "table3")
