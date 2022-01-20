package dev.beneath.cdc.postgres;

import dev.beneath.client.BeneathClient;

public class App {
    public static void main(String[] args) {
        BeneathClient client = new BeneathClient(CdcConfig.BENEATH_SECRET, false, CdcConfig.DEFAULT_WRITE_DELAY_MS);
        BeneathDebeziumEngine engine = new BeneathDebeziumEngine(client);
        engine.start();
    }
}
