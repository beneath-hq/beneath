# to start the CDC, take a snapshot of the database, and write everything to Beneath

# use pg_export_snapshot() to set the snapshot of the database
# use the same snapshot in create_logical_replication_slot(), so I can ensure that no data updates have happened in between

# Because
# a) "The snapshot is available for import only until the end of the transaction that exported it"
# and b)

# I think I need to FIRST create the replication_slot (which will open a long-running transaction),
# THEN take the snapshot (which should be a quick transaction) w/ SET TRANSACTION SNAPSHOT = [id from pg_export_snapshot()]