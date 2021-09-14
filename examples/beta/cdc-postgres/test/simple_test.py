import os
import sys
import inspect

# add parent directory to path, so I can import from the cdc.py module
currentdir = os.path.dirname(os.path.abspath(inspect.getfile(inspect.currentframe())))
parentdir = os.path.dirname(currentdir)
sys.path.insert(0, parentdir)

from cdc import connect_to_source_db

if __name__ == "__main__":
    conn = connect_to_source_db()
    cursor = conn.cursor()
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
    cursor.execute(
        f"""
    CREATE TABLE actions (a integer primary key);
    INSERT INTO actions (a) VALUES(1);
    UPDATE actions SET a = 2 WHERE a = 1;
    DELETE FROM actions WHERE a = 2;
    TRUNCATE TABLE actions;
    """
    )

    cursor.execute(
        "SELECT data FROM pg_logical_slot_get_changes('test_slot', NULL, NULL);"
    )  # for production
    changes = cursor.fetchall()
    print(changes)
    print(len(changes))
