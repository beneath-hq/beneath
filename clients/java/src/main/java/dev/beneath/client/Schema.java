package dev.beneath.client;

import java.io.ByteArrayOutputStream;
import java.io.IOException;
import java.util.Map;
import java.util.Map.Entry;

import com.google.protobuf.ByteString;

import org.apache.avro.generic.GenericDatumReader;
import org.apache.avro.generic.GenericDatumWriter;
import org.apache.avro.generic.GenericRecord;
import org.apache.avro.io.BinaryEncoder;
import org.apache.avro.io.DatumReader;
import org.apache.avro.io.DatumWriter;
import org.apache.avro.io.Decoder;
import org.apache.avro.io.DecoderFactory;
import org.apache.avro.io.EncoderFactory;

/**
 * Represents a table's parsed Avro schema
 */
public class Schema {
  public String avro;
  public org.apache.avro.Schema parsedAvro;

  public Schema(String avro) {
    this.avro = avro;
    this.parsedAvro = new org.apache.avro.Schema.Parser().parse(avro);
  }

  // TODO: what's the best way to return a tuple?
  public Entry<Record, Integer> recordToPb(GenericRecord record) {
    byte[] avro = this.encodeAvro(record);
    Integer timestamp = extractRecordTimestamp(record);
    Record pb = Record.newBuilder().setAvroData(ByteString.copyFrom(avro)).setTimestamp(timestamp).build();
    return Map.entry(pb, pb.getSerializedSize());
  }

  public GenericRecord pbToRecord(Record pb) {
    GenericRecord record = this.decodeAvro(pb.getAvroData().toByteArray());
    // record.put("@meta.timestamp", "not yet implemented");
    return record;
  }

  private byte[] encodeAvro(GenericRecord record) {
    ByteArrayOutputStream out = new ByteArrayOutputStream();
    BinaryEncoder encoder = EncoderFactory.get().binaryEncoder(out, null);
    DatumWriter<GenericRecord> writer = new GenericDatumWriter<GenericRecord>(this.parsedAvro);
    try {
      writer.write(record, encoder);
      encoder.flush();
      out.close();
    } catch (IOException e) {
      throw new RuntimeException(e);
    }
    byte[] serializedBytes = out.toByteArray();
    return serializedBytes;
  }

  private GenericRecord decodeAvro(byte[] data) {
    DatumReader<GenericRecord> reader = new GenericDatumReader<GenericRecord>(this.parsedAvro);
    Decoder decoder = DecoderFactory.get().binaryDecoder(data, null);
    GenericRecord record;
    try {
      record = reader.read(null, decoder);
    } catch (IOException e) {
      throw new RuntimeException(e);
    }
    return record;
  }

  private static Integer extractRecordTimestamp(GenericRecord record) {
    // TODO: actually check the record for a timestamp
    return 0; // 0 tells the server to set timestamp to its current time
  }
}
