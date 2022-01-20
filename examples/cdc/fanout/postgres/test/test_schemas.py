import os, inspect, sys

# add parent directory to path, so I can import from the main.py module
currentdir = os.path.dirname(os.path.abspath(inspect.getfile(inspect.currentframe())))
parentdir = os.path.dirname(currentdir)
sys.path.insert(0, parentdir)

from main import connect_to_source_db
from operate_on_source_db import create_table_compound_pk, drop_table
from schemas import get_schema

if __name__ == "__main__":
    cursor = connect_to_source_db()
    drop_table(cursor, "table1")
    create_table_compound_pk(cursor, "table1")

    schema = get_schema(cursor, "table1")
    print(schema)