package dev.beneath;

import java.util.Properties;
import java.util.concurrent.ExecutorService;
import java.util.concurrent.Executors;

import org.slf4j.Logger;
import org.slf4j.LoggerFactory;

import io.debezium.engine.ChangeEvent;
import io.debezium.engine.DebeziumEngine;
import io.debezium.engine.format.Json;
import io.debezium.engine.spi.OffsetCommitPolicy;

public class BeneathDebeziumEngine {
    private static final Logger LOGGER = LoggerFactory.getLogger(BeneathDebeziumEngine.class);
    DebeziumEngine<ChangeEvent<String, String>> engine;

    BeneathDebeziumEngine() {
    };

    public void start() {
        Properties properties = getProperties();
        OffsetCommitPolicy offsetPolicy = new OffsetCommitPolicy.AlwaysCommitOffsetPolicy();

        // Create the engine with this configuration ...
        engine = DebeziumEngine.create(Json.class).using(properties).using(offsetPolicy)
                .notifying((records, committer) -> {
                    for (ChangeEvent<String, String> r : records) {
                        LOGGER.info("Key = '" + r.key() + "' value = '" + r.value() + "'");
                        committer.markProcessed(r);
                    }
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
        props.setProperty("name", "postgres-debezium-connector");
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

        // TODO: think about schemas. This setting disables schemas from getting
        // included in each record (which would significantly increase amount of data
        // transferred)
        props.setProperty("converter.schemas.enable", "False");
        return props;
    }
}
