package dev.beneath;

import dev.beneath.client.BeneathClient;
import dev.beneath.client.Config;

public class App {
    public static void main(String[] args) {
        BeneathClient client = new BeneathClient(DebeziumConfig.BENEATH_SECRET, false, Config.DEFAULT_WRITE_DELAY_MS);
        BeneathDebeziumEngine engine = new BeneathDebeziumEngine(client);
        engine.start();
    }
}
