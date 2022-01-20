This project connects a Postgres database to Beneath.

## 1. Fill out config.yaml

All source tables must have primary keys.

## 2. Configure your Postgres server

Someone with the requisite permissions should run these commands.

To connect to the database:
```bash
psql -d source_db
```

### wal_level
From the Postgres docs: "wal_level determines how much information is written to the WAL. The default value is `replica`, which writes enough data to support WAL archiving and replication, including running read-only queries on a standby server. `logical` adds information necessary to support logical decoding. This parameter can only be set at server start."

In psql command prompt:
```sql
ALTER SYSTEM SET wal_level=logical;
```
Then restart the Postgres server.

Confirm the setting is correct with:
```sql
SHOW wal_level;
```

### publication
Only necessary if using the `pgoutput` plugin.

From the Postgres docs: "A publication is essentially a group of tables whose data changes are intended to be replicated through logical replication."

You must create the publication *before* creating the replication slot.

```sql
CREATE PUBLICATION beneath_​publication FOR ALL TABLES;
```

### replication_slot
From the Postgres docs: "In the context of logical replication, a slot represents a stream of changes that can be replayed to a client in the order they were made on the origin server. Each slot streams a sequence of changes from a single database."


In psql command prompt:
```sql
SELECT pg_create_logical_replication_slot('beneath_replication_slot', 'wal2json');
```

Important: make sure to clean up unused replication slots. See the discussion [here](https://techcommunity.microsoft.com/t5/azure-database-for-postgresql/change-data-capture-in-postgres-how-to-use-logical-decoding-and/ba-p/1396421).

## 3. Run the Beneath pipeline
```python
python main.py stage username/project/service
python main.py run username/project/service
```

To clean up the Beneath resources:
```python
python main.py teardown username/project/service
```

## Resources
- Postgres logical decoding docs: https://www.postgresql.org/docs/current/logicaldecoding-explanation.html#LOGICALDECODING-REPLICATION-SLOTS
- Azure's explanation: https://techcommunity.microsoft.com/t5/azure-database-for-postgresql/change-data-capture-in-postgres-how-to-use-logical-decoding-and/ba-p/1396421
- GCP's tutorial: https://cloud.google.com/sql/docs/postgres/replication/configure-logical-replication#receiving-decoded-wal-changes-for-change-data-capture
- AWS: https://docs.aws.amazon.com/AmazonRDS/latest/UserGuide/CHAP_PostgreSQL.html#PostgreSQL.Concepts.General.FeatureSupport.LogicalReplication
- Debezium's Postgres docs: https://debezium.io/documentation/reference/1.0/connectors/postgresql.html
- Embedded Debezium docs: https://debezium.io/documentation/reference/1.6/development/engine.html
- Airbyte's approach: https://opensource.com/article/21/8/database-replication-open-source
- Airbyte's demo: https://www.youtube.com/watch?v=NMODvLgZvuE&ab_channel=Airbyte
- Fivetran's pgoutput configuration docs: https://fivetran.com/docs/databases/postgresql/setup-guide#logicalreplicationwiththepluginprivatepreview