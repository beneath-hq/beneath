package dev.beneath;

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
import org.slf4j.Logger;
import org.slf4j.LoggerFactory;

import dev.beneath.client.BeneathClient;
import dev.beneath.client.Checkpointer;
import dev.beneath.client.Table;
import dev.beneath.client.TableInstance;
import dev.beneath.client.utils.JsonUtils;
import dev.beneath.client.utils.TableIdentifier;
import dev.beneath.type.TableSchemaKind;
import io.debezium.engine.ChangeEvent;
import io.debezium.engine.DebeziumEngine;
import io.debezium.engine.format.Json;
import io.debezium.engine.spi.OffsetCommitPolicy;

public class BeneathDebeziumEngine {
    private static final Logger LOGGER = LoggerFactory.getLogger(BeneathDebeziumEngine.class);
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
    private BeneathClient client;
    private Checkpointer checkpointer;
    private TableIdentifier rootTableIdentifier;
    private String tableDescription;
    private Schema avroSchema;
    private DebeziumEngine<ChangeEvent<String, String>> engine;
    private TableInstance instance;
    private boolean create;

    BeneathDebeziumEngine(BeneathClient client) {
        this.client = client;
        this.checkpointer = client.checkpointer(DebeziumConfig.BENEATH_PROJECT_PATH);
        this.rootTableIdentifier = TableIdentifier.fromPath(DebeziumConfig.BENEATH_DEBEZIUM_ROOT_TABLE_PATH);
        this.tableDescription = "Records streamed from Postgres through Debezium";
        this.create = true;
    };

    public void start() {
        Properties properties = getProperties();
        OffsetCommitPolicy offsetPolicy = new OffsetCommitPolicy.AlwaysCommitOffsetPolicy();

        this.client.start();
        stageBeneathTable();

        // Create the engine with this configuration ...
        engine = DebeziumEngine.create(Json.class).using(properties).using(offsetPolicy)
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

    private Properties getProperties() {
        final Properties props = new Properties();

        /* engine properties */
        props.setProperty("name", "debezium-postgres-connector");
        props.setProperty("connector.class", "io.debezium.connector.postgresql.PostgresConnector");
        props.setProperty("offset.storage", "dev.beneath.BeneathOffsetBackingStore");
        props.setProperty("offset.flush.interval.ms", "1000");

        /* connector properties */
        props.setProperty("database.hostname", DebeziumConfig.DATABASE_HOSTNAME);
        props.setProperty("database.port", DebeziumConfig.DATABASE_PORT);
        props.setProperty("database.user", DebeziumConfig.DATABASE_USER);
        props.setProperty("database.password", DebeziumConfig.DATABASE_PASSWORD);
        props.setProperty("database.dbname", DebeziumConfig.DATABASE_DBNAME);
        props.setProperty("table.include.list", DebeziumConfig.TABLE_INCLUDE_LIST);
        props.setProperty("plugin.name", "pgoutput");
        props.setProperty("database.server.name", DebeziumConfig.DATABASE_SERVER_NAME);

        // future optimization: set this to false to significantly decrease inbound
        // bandwidth costs; likely requires optimistically writing to Beneath and, on
        // failure, querying the source db's data model
        props.setProperty("converter.schemas.enable", "True");

        return props;
    }

    private void stageBeneathTable() {
        Table table;
        if (this.create || this.client.dry) {
            table = this.client.createTable(this.rootTableIdentifier.toString(), DEBEZIUM_PG_TABLE_SCHEMA,
                    this.tableDescription, false, true, true, 0, 0, 0, 0, TableSchemaKind.GRAPHQL, "", true);
        } else {
            table = this.client.findTable(this.rootTableIdentifier.toString());
        }
        if (table.primaryInstance == null) {
            throw new RuntimeException("Expected the debezium table to have a primary instance");
        }
        this.instance = table.primaryInstance;
        this.avroSchema = new Schema.Parser().parse(this.instance.table.schema.parsedAvro.toString());

        LOGGER.info("Using '{}' (version {}) for debezium", this.rootTableIdentifier.toString(), this.instance.version);
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
