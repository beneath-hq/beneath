import psycopg2
import yaml

if __name__ == "__main__":
    # load config
    with open(".development.yaml", "r") as ymlfile:
        cfg = yaml.safe_load(ymlfile)

    # connect to "postgres" db
    # connecting to "postgres" db allows me to drop & create other dbs
    conn = psycopg2.connect(
        database="postgres",
        user=cfg["postgres"]["username"],
        password=cfg["postgres"]["password"],
        host=cfg["postgres"]["host"],
        port="5432",
    )
    conn.autocommit = True
    cursor = conn.cursor()

    # create a test database
    cursor.execute("DROP DATABASE IF EXISTS testdb")
    cursor.execute("CREATE DATABASE testdb")
    print("Created test database.")

    # close "postgres", open "testdb"
    conn.close()
    conn = psycopg2.connect(
        database="testdb",
        user=cfg["postgres"]["username"],
        password=cfg["postgres"]["password"],
        host=cfg["postgres"]["host"],
        port="5432",
    )
    conn.autocommit = True
    cursor = conn.cursor()

    # create user table
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
    print("Created user table.")

    # populate with data
    cursor.execute(
        """
    INSERT INTO users VALUES (1, 'john@test.com', False, now(), now());
      """
    )
    print("Inserted a dummy record.")

    # TODO: insert random records over a time period
    # TODO: test updates and deletes, too

    # close testdb connection
    conn.close()