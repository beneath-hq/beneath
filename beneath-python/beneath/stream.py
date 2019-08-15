import sys
import io
import grpc
import uuid
import json
import time
import pandas as pd
from fastavro import schemaless_writer, schemaless_reader, reader, parse_schema
from beneath import config
from beneath.proto import gateway_pb2_grpc
from beneath.proto import gateway_pb2
from beneath.proto import engine_pb2

class Stream:
  """
  Stream enables read and write operations on Beneath streams
  """

  def __init__(self, client, project_name, stream_name, schema, avro_schema, key_fields, batch, current_instance_id):
    """
    Args:
      client (Client): Connection to Beneath.
      project_name (str): Project name.
      stream_name (str): Stream name.
      schema (str): The stream's schema.
      avro_schema (str): Avro representation of the stream's schema.
      batch (bool): Whether writes overwrite or append to the stream.
      current_instance_id (UUID): ID of current stream instance for data reads.
    """
    self.client = client
    self.project_name = project_name
    self.stream_name = stream_name
    self.schema = schema
    self.avro_schema = parse_schema(json.loads(avro_schema))
    self.key_fields = key_fields
    self.batch = batch
    self.current_instance_id = current_instance_id


  def __getstate__(self):
    return {
        "client": self.client,
        "project_name": self.project_name,
        "stream_name": self.stream_name,
        "schema": self.schema,
        "avro_schema": self.avro_schema,
        "key_fields": self.key_fields,
        "batch": self.batch,
        "current_instance_id": self.current_instance_id,
    }


  def __setstate__(self, obj):
    self.client = obj["client"]
    self.project_name = obj["project_name"]
    self.stream_name = obj["stream_name"]
    self.schema = obj["schema"]
    self.avro_schema = obj["avro_schema"]
    self.key_fields = obj["key_fields"]
    self.batch = obj["batch"]
    self.current_instance_id = obj["current_instance_id"]


  def read(self, where=None, max_rows=None, max_megabytes=None, instance_id=None):
    instance_id = self._instance_id_or_default(instance_id)
    where = self._parse_where(where)
    
    max_rows = max_rows if max_rows else sys.maxsize
    max_megabytes = max_megabytes if max_megabytes else config.MAX_READ_MB
    max_bytes = max_megabytes * (2**20)

    records = []
    rows_loaded = 0
    bytes_loaded = 0

    while rows_loaded < max_rows and bytes_loaded < max_bytes:
      after = None
      if len(records) > 0:
        after = { field: records[-1][field] for field in self.key_fields }
      
      limit = min(max_rows - rows_loaded, config.READ_BATCH_SIZE)

      batch = self.client.read_batch(instance_id, where, limit, after)
      if len(batch) == 0:
        break

      for record in batch:
        records.append(self._decode_avro(record.avro_data))
        rows_loaded += 1
        bytes_loaded += len(record.avro_data)
    
    return pd.DataFrame(records)


  def write_records(self, instance_id, records, timestamp=None):  # should I be multiprocessing this for loop?
    new_records = [None]*len(records)

    # ensure each record is a dict
    for i, record in enumerate(records):
      if not isinstance(record, dict):
        print(record)
        raise TypeError("record must be a dict")

      # encode avro
      encoded_data = self._encode_avro(record)
      if timestamp is None:
        timestamp = int(round(time.time() * 1000))
      new_records[i] = engine_pb2.Record(
          avro_data=encoded_data, timestamp=timestamp)

    # gRPC WriteRecords to gateway
    response = self.client.stub.WriteRecords(
        engine_pb2.WriteRecordsRequest(instance_id=instance_id.bytes, records=new_records), metadata=self.client.request_metadata)
    return response


  def bigquery_name(self, table=False):
    return "{}.{}.{}{}".format(
      config.BIGQUERY_PROJECT,
      self.project_name.replace("-", "_"),
      self.stream_name.replace("-", "_"),
      "_{}".format(self.current_instance_id.hex[0:6]) if table else ""
    )


  def _instance_id_or_default(self, instance_id):
    # instance ID defaults to current_instance_id
    instance_id = instance_id if instance_id else self.current_instance_id
    if instance_id is None:
      # for batch streams, current_instance_id may be null
      raise Exception(
          "Cannot query stream because instance ID is null"
          " (Is it a batch stream that has not yet finished its first load?)"
      )
    return instance_id


  def _decode_avro(self, data):
    reader = io.BytesIO(data)
    record = schemaless_reader(reader, self.avro_schema)
    reader.close()
    return record


  def _encode_avro(self, record):
    writer = io.BytesIO()
    schemaless_writer(writer, self.avro_schema, record)
    result = writer.getvalue()
    writer.close()
    return result


  def _parse_where(self, where):
    if isinstance(where, str):
      return where
    elif isinstance(where, dict):
      return json.dumps(where)
    else:
      raise TypeError("expected json string or dict for parameter 'where'")
