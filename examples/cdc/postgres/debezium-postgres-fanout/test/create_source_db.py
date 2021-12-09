import psycopg2
import yaml

if __name__ == "__main__":
    # load config
    with open(".development.yaml", "r") as ymlfile:
        config = yaml.safe_load(ymlfile)

    # connect to "postgres" db
    # connecting to "postgres" db allows me to drop & create other dbs
    conn = psycopg2.connect(
        database="postgres",
        user=config["postgres"]["username"],
        password=config["postgres"]["password"],
        host=config["postgres"]["host"],
        port="5432",
    )
    conn.autocommit = True
    cursor = conn.cursor()

    # create a test database
    cursor.execute(f"DROP DATABASE IF EXISTS {config['postgres']['database']}")
    cursor.execute(f"CREATE DATABASE {config['postgres']['database']}")
    print("Created test database.")

    # close connection
    conn.close()