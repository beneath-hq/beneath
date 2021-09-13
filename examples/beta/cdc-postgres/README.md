# Config
TODO: add motivation for config

## connect to the test database
```bash
psql -d testdb
```

## wal_level
TODO: explain what the wal_level is

In psql command prompt:
```sql
ALTER SYSTEM SET wal_level=logical;
```
Then restart the Postgres server.

Confirm the setting is correct with:
```sql
SHOW wal_level;
```

## replication_slot
TODO: explain what a replication slot is

In psql command prompt:
```sql
SELECT pg_create_logical_replication_slot('test_slot', 'wal2json');
```