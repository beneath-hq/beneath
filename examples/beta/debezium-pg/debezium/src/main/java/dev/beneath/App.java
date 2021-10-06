package dev.beneath;

import java.util.Properties;
import java.util.concurrent.ExecutorService;
import java.util.concurrent.Executors;
import java.util.concurrent.TimeUnit;

import io.debezium.engine.ChangeEvent;
import io.debezium.engine.DebeziumEngine;
import io.debezium.engine.format.Json;

public class App {
    public static void main(String[] args) {
        final Properties props = new Properties();
        props.setProperty("connector.class", "io.debezium.connector.postgresql.PostgresConnector");
        props.setProperty("name", "postgres-debezium-connector");

        // TODO: implement Beneath offset storage
        props.setProperty("offset.storage", "org.apache.kafka.connect.storage.FileOffsetBackingStore");
        props.setProperty("offset.storage.file.filename", "/tmp/offsets.dat");
        props.setProperty("offset.flush.interval.ms", "60000");

        /* begin connector properties */
        // TODO: move config to `config.properties` file
        props.setProperty("database.hostname", "host.docker.internal");
        props.setProperty("database.port", "5432");
        props.setProperty("database.user", "ericgreen");
        props.setProperty("database.password", "");
        props.setProperty("database.dbname", "testdb");
        props.setProperty("table.include.list", "public.table1, public.table2, public.table3");
        props.setProperty("plugin.name", "pgoutput");
        props.setProperty("database.server.name", "test");

        // TODO: think about schemas. This setting disables schemas from getting
        // included in each record (which would significantly increase amount of data
        // transferred)
        props.setProperty("converter.schemas.enable", "False");

        // Create the engine with this configuration ...
        final DebeziumEngine<ChangeEvent<String, String>> engine = DebeziumEngine.create(Json.class).using(props)
                .notifying((records, committer) -> {
                    for (ChangeEvent<String, String> r : records) {
                        System.out.println("Key = '" + r.key() + "' value = '" + r.value() + "'");

                        // TODO: write to Beneath

                        committer.markProcessed(r);
                    }
                }).build();

        // Run the engine asynchronously ...
        ExecutorService executor = Executors.newSingleThreadExecutor();
        executor.execute(engine);

        // Do something else or wait for a signal or an event
        try {
            while (!executor.awaitTermination(10, TimeUnit.SECONDS)) {
                System.out.println("Waiting another 10 seconds for the embedded engine to complete");
            }
        } catch (InterruptedException e) {
            Thread.currentThread().interrupt();
        }

        // Engine is stopped when the main code is finished
    }
}
