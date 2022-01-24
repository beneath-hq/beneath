package dev.beneath.cdc.postgres;

import java.util.ArrayList;
import java.util.List;
import java.util.Optional;
import java.util.Properties;
import java.util.concurrent.ExecutorService;
import java.util.concurrent.Executors;

import com.fasterxml.jackson.databind.JsonNode;

import org.apache.avro.Schema;
import org.apache.avro.generic.GenericRecord;
import org.apache.avro.generic.GenericRecordBuilder;

import dev.beneath.cdc.common.DbzRecordSchema;
import dev.beneath.client.BeneathClient;
import dev.beneath.client.Checkpointer;
import dev.beneath.client.Table;
import dev.beneath.client.TableInstance;
import dev.beneath.client.type.TableSchemaKind;
import dev.beneath.client.utils.JsonUtils;
import dev.beneath.client.utils.TableIdentifier;
import io.debezium.engine.ChangeEvent;
import io.debezium.engine.DebeziumEngine;
import io.debezium.engine.format.Json;
import io.debezium.engine.spi.OffsetCommitPolicy;

public class BeneathDebeziumEngine {
    // each connector generates a unique "source" field, so this schema is unique to
    // the postgres connector
    private static final String DEBEZIUM_PG_TABLE_SCHEMA = """
            type Change @schema {
                source_db: String! @key
                source_schema: String! @key
                source_table: String! @key
                source_ts_ms: Timestamp! @key
                source_connector: String!
                source_version: String!
                source_name: String!
                source_snapshot: Boolean!
                source_sequence: String!
                source_tx_id: Int!
                source_lsn: Int!
                source_xmin: String
                op: String!
                ts_ms: Timestamp!
                transaction: String
                before: String
                after: String!
            }
            """;
    private static final String PROJECT_DESCRIPTION = "Automatically created by the Beneath CDC Service.";
    private static final String PROJECT_SITE = "https://github.com/beneath-hq/beneath/tree/master/examples/cdc";
    private static final String TABLE_DESCRIPTION = "Records streamed from Postgres through Debezium";

    private BeneathClient client;
    private Checkpointer checkpointer;
    private TableIdentifier rootTableIdentifier;
    private Schema avroSchema;
    private TableInstance instance;
    private DebeziumEngine<ChangeEvent<String, String>> engine;

    public BeneathDebeziumEngine(BeneathClient client) {
        this.client = client;
        this.checkpointer = client.checkpointer(CdcConfig.BENEATH_USERNAME + "/" + CdcConfig.BENEATH_PROJECT_NAME);
        this.rootTableIdentifier = new TableIdentifier(CdcConfig.BENEATH_USERNAME, CdcConfig.BENEATH_PROJECT_NAME,
                CdcConfig.BENEATH_DEBEZIUM_ROOT_TABLE_NAME);
    };

    public void start() {
        stageBeneathResources();
        Properties dbzProperties = getDebeziumProperties();
        OffsetCommitPolicy offsetPolicy = new OffsetCommitPolicy.AlwaysCommitOffsetPolicy();

        // Create the engine with this configuration ...
        engine = DebeziumEngine.create(Json.class).using(dbzProperties).using(offsetPolicy)
                .notifying((dbzRecords, committer) -> {
                    List<GenericRecord> beneathRecords = new ArrayList<GenericRecord>();
                    for (ChangeEvent<String, String> r : dbzRecords) {
                        // load json
                        JsonNode dbzRecordKey = JsonUtils.deserialize(r.key(), JsonNode.class);
                        JsonNode dbzRecordValue = JsonUtils.deserialize(r.value(), JsonNode.class);

                        // if there's a new schema, write it to Beneath
                        DbzRecordSchema dbzRecordSchema = getDbzRecordSchema(dbzRecordKey, dbzRecordValue);
                        String keyForCheckpointedSchema = makeKeyForCheckpointedSchema(dbzRecordValue);
                        Optional<DbzRecordSchema> checkpointedSchema = getCheckpointedSchema(keyForCheckpointedSchema);
                        if (checkpointedSchema.isEmpty() || !dbzRecordSchema.equals(checkpointedSchema.get())) {
                            setCheckpointedSchema(keyForCheckpointedSchema, dbzRecordSchema);
                        }

                        // construct record
                        GenericRecord beneathRecord = createBeneathRecord(dbzRecordValue);
                        beneathRecords.add(beneathRecord);
                        committer.markProcessed(r); // Q: does this do anything?
                    }
                    this.instance.write(beneathRecords);
                    committer.markBatchFinished(); // flush offsets
                }).build();

        // Run the engine asynchronously ...
        ExecutorService executor = Executors.newSingleThreadExecutor();
        executor.execute(engine);

        // Engine is stopped when the main code is finished
    }

    private void stageBeneathResources() {
        this.client.stageProject(CdcConfig.BENEATH_USERNAME, CdcConfig.BENEATH_PROJECT_NAME, false,
                BeneathDebeziumEngine.PROJECT_DESCRIPTION, BeneathDebeziumEngine.PROJECT_SITE);
        Table table = this.client.stageTable(this.rootTableIdentifier.toString(), DEBEZIUM_PG_TABLE_SCHEMA,
                BeneathDebeziumEngine.TABLE_DESCRIPTION, false, true, true, 0, 0, 0, 0, TableSchemaKind.GRAPHQL, "",
                true);
        this.instance = table.primaryInstance;
        this.avroSchema = new Schema.Parser().parse(this.instance.table.schema.parsedAvro.toString());
        this.client.start(); // starts the checkpointer
    }

    private Properties getDebeziumProperties() {
        final Properties props = new Properties();

        /* engine properties */
        props.setProperty("name", "debezium-postgres-connector");
        props.setProperty("connector.class", "io.debezium.connector.postgresql.PostgresConnector");
        props.setProperty("offset.storage", "dev.beneath.cdc.common.BeneathOffsetBackingStore");
        props.setProperty("offset.flush.interval.ms", "1000");

        /* connector properties */
        props.setProperty("database.hostname", CdcConfig.DATABASE_HOSTNAME);
        props.setProperty("database.port", CdcConfig.DATABASE_PORT);
        props.setProperty("database.user", CdcConfig.DATABASE_USER);
        props.setProperty("database.password", CdcConfig.DATABASE_PASSWORD);
        props.setProperty("database.dbname", CdcConfig.DATABASE_DBNAME);
        props.setProperty("table.include.list", CdcConfig.TABLE_INCLUDE_LIST);
        props.setProperty("plugin.name", "pgoutput");
        props.setProperty("database.server.name", CdcConfig.DATABASE_SERVER_NAME);

        // future optimization: set this to false to significantly decrease inbound
        // bandwidth costs; likely requires optimistically writing to Beneath and, on
        // failure, querying the source db's data model
        props.setProperty("converter.schemas.enable", "True");

        return props;
    }

    private DbzRecordSchema getDbzRecordSchema(JsonNode key, JsonNode value) {
        String keySchema = key.get("schema").get("fields").toString();
        String valueSchema = value.get("schema").get("fields").get(1).get("fields").toString(); // "1" = "after"
        return new DbzRecordSchema(keySchema, valueSchema);
    }

    private String makeKeyForCheckpointedSchema(JsonNode value) {
        String pgDatabase = value.get("payload").get("source").get("db").asText();
        String pgSchema = value.get("payload").get("source").get("schema").asText();
        String pgTable = value.get("payload").get("source").get("table").asText();
        return pgDatabase + ":" + pgSchema + ":" + pgTable + ":" + "schema";
    }

    private Optional<DbzRecordSchema> getCheckpointedSchema(String keyForCheckpointedSchema) {
        String checkpointedSchemaAsString = (String) this.checkpointer.get(keyForCheckpointedSchema, null);
        if (checkpointedSchemaAsString == null) {
            return Optional.empty();
        }
        return Optional.of(JsonUtils.deserialize(checkpointedSchemaAsString, DbzRecordSchema.class));
    }

    private void setCheckpointedSchema(String keyForCheckpointedSchema, DbzRecordSchema dbzRecordSchema) {
        String recordSchemaAsString = JsonUtils.serialize(dbzRecordSchema);
        this.checkpointer.set(keyForCheckpointedSchema, recordSchemaAsString);
    }

    private GenericRecord createBeneathRecord(JsonNode value) {
        String sourceDb = value.get("payload").get("source").get("db").asText();
        String sourceSchema = value.get("payload").get("source").get("schema").asText();
        String sourceTable = value.get("payload").get("source").get("table").asText();
        Long sourceTsMs = value.get("payload").get("source").get("ts_ms").asLong();
        String sourceConnector = value.get("payload").get("source").get("connector").asText();
        String sourceVersion = value.get("payload").get("source").get("version").asText();
        String sourceName = value.get("payload").get("source").get("name").asText();
        Boolean sourceSnapshot = value.get("payload").get("source").get("snapshot").asBoolean();
        String sourceSequence = value.get("payload").get("source").get("sequence").asText();
        Long sourceTxId = value.get("payload").get("source").get("txId").asLong();
        Long sourceLsn = value.get("payload").get("source").get("lsn").asLong();
        String sourceXmin = value.get("payload").get("source").get("xmin").asText();
        String op = value.get("payload").get("op").asText();
        Long tsMs = value.get("payload").get("ts_ms").asLong();
        String transaction = value.get("payload").get("transaction").asText();
        String before = value.get("payload").get("before").toString() == "null" ? null
                : value.get("payload").get("before").toString();
        String after = value.get("payload").get("after").toString() == "null" ? null
                : value.get("payload").get("after").toString();
        GenericRecord beneathRecord = new GenericRecordBuilder(avroSchema).set("source_db", sourceDb)
                .set("source_schema", sourceSchema).set("source_table", sourceTable).set("source_ts_ms", sourceTsMs)
                .set("source_connector", sourceConnector).set("source_version", sourceVersion)
                .set("source_name", sourceName).set("source_snapshot", sourceSnapshot)
                .set("source_sequence", sourceSequence).set("source_tx_id", sourceTxId).set("source_lsn", sourceLsn)
                .set("source_xmin", sourceXmin).set("op", op).set("ts_ms", tsMs).set("transaction", transaction)
                .set("before", before).set("after", after).build();
        return beneathRecord;
    }
}
