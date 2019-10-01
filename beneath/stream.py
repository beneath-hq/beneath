import io
from datetime import datetime, timezone
import json
import sys
import warnings

from fastavro import schemaless_writer, schemaless_reader, parse_schema
import pandas as pd

from beneath import config
from beneath.proto import engine_pb2

class Stream:
  """
  Stream enables read and write operations on Beneath streams
  """

  def __init__(self, client, project_name, stream_name, schema, key_fields, avro_schema, batch, current_instance_id):
    """
    Args:
      client (Client): Connection to Beneath.
      project_name (str): Project name.
      stream_name (str): Stream name.
      schema (str): The stream's schema.
      key_fields (list(str)): Fields that make up the stream's primary key.
      avro_schema (str): Avro representation of the stream's schema.
      batch (bool): Whether writes overwrite or append to the stream.
      current_instance_id (UUID): ID of current stream instance for data reads.
    """
    self.client = client
    self.project_name = project_name
    self.stream_name = stream_name
    self.schema = schema
    self.key_fields = key_fields
    self.avro_schema = parse_schema(json.loads(avro_schema))
    self.batch = batch
    self.current_instance_id = current_instance_id


  def __getstate__(self):
    return {
      'client': self.client,
      'project_name': self.project_name,
      'stream_name': self.stream_name,
      'schema': self.schema,
      'key_fields': self.key_fields,
      'avro_schema': self.avro_schema,
      'batch': self.batch,
      'current_instance_id': self.current_instance_id,
    }


  def __setstate__(self, obj):
    self.client = obj['client']
    self.project_name = obj['project_name']
    self.stream_name = obj['stream_name']
    self.schema = obj['schema']
    self.key_fields = obj['key_fields']
    self.avro_schema = obj['avro_schema']
    self.batch = obj['batch']
    self.current_instance_id = obj['current_instance_id']


  def write(self, records, instance_id=None):
    instance_id = self._instance_id_or_default(instance_id)
    encoded_records = (self._encode_record(record) for record in records)
    self.client.write_batch(instance_id, encoded_records)


  def read(self, where=None, max_rows=None, max_megabytes=None, instance_id=None, to_dataframe=True):
    instance_id = self._instance_id_or_default(instance_id)
    where = self._parse_where(where)

    def read(limit, after):
      return self.client.read_batch(instance_id=instance_id, where=where, limit=limit, after=after)

    return self._constrained_read(read, max_rows=max_rows, max_megabytes=max_megabytes, to_dataframe=to_dataframe)


  def latest(self, max_rows=None, max_megabytes=None, instance_id=None, to_dataframe=True):
    instance_id = self._instance_id_or_default(instance_id)

    def read(limit, after):
      return self.client.read_latest_batch(instance_id=instance_id, limit=limit, after=after)

    return self._constrained_read(read, max_rows=max_rows, max_megabytes=max_megabytes, to_dataframe=to_dataframe)


  def _constrained_read(self, read_fn, max_rows=None, max_megabytes=None, to_dataframe=True):
    # adding 1 to turn <= into <
    max_rows = max_rows + 1 if max_rows else sys.maxsize
    max_megabytes = max_megabytes if max_megabytes else config.MAX_READ_MB
    max_bytes = max_megabytes * (2**20) + 1

    records = []
    complete = False
    rows_loaded = 0
    bytes_loaded = 0

    while rows_loaded < max_rows and bytes_loaded < max_bytes:
      after = None
      if len(records) > 0:
        after = {field: records[-1][field] for field in self.key_fields}
        after = json.dumps(after, default=self._json_encode)

      limit = min(max_rows - rows_loaded, config.READ_BATCH_SIZE)
      if limit == 0:
        break

      batch = read_fn(limit=limit, after=after)
      if len(batch) == 0:
        complete = True
        break

      for record in batch:
        records.append(self._decode_avro(record.avro_data))
        rows_loaded += 1
        bytes_loaded += len(record.avro_data)
        if bytes_loaded >= max_bytes:
          break

      if len(batch) < limit:
        complete = True
        break

    if not complete:
      # Jupyter doesn't always display warnings, so also print
      # Note: rememberr max_rows and max_bytes are +1
      if rows_loaded >= max_rows:
        err = "Stopped loading because stream length exceeds max_rows={}".format(max_rows - 1)
        print(err)
        warnings.warn(err)
      elif bytes_loaded >= max_bytes:
        err = "Stopped loading because download size exceeds max_megabytes={}MB".format(max_megabytes - 1)
        print(err)
        warnings.warn(err)

    if to_dataframe:
      return pd.DataFrame(records, columns=self._columns)
    return records


  @property
  def _columns(self):
    return [field["name"] for field in self.avro_schema["fields"]]


  @property
  def bigquery_table(self):
    return self.get_bigquery_table()


  def get_bigquery_table(self, view=True, instance_id=None):
    uid = self._instance_id_or_default(instance_id)
    return "{}.{}.{}{}".format(
      config.BIGQUERY_PROJECT,
      self.project_name.replace("-", "_"),
      self.stream_name.replace("-", "_"),
      "" if view else "_{}".format(uid.hex[0:8])
    )


  def _encode_record(self, record):
    if not isinstance(record, dict):
      raise TypeError("write error: record must be a dict, got {}".format(record))
    timestamp = self._extract_record_timestamp(record)
    avro = self._encode_avro(record)
    return engine_pb2.Record(avro_data=avro, timestamp=timestamp)


  @classmethod
  def _extract_record_timestamp(cls, record):
    if ("@meta" in record) and ("timestamp" in record["@meta"]):
      return cls._timestamp_to_ms(record["@meta"]["timestamp"])
    return 0 # 0 tells the server to set timestamp to its current time


  def _instance_id_or_default(self, instance_id):
    instance_id = instance_id if instance_id else self.current_instance_id
    if instance_id is None:
      # for batch streams, current_instance_id may be null
      raise Exception(
        "Cannot query stream because instance ID is null."
        " (Is it a batch stream that has not yet finished its first load?)"
        " If a new instance has just been committed, reload the stream object."
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


  @classmethod
  def _parse_where(cls, where):
    if where is None:
      return ""
    if isinstance(where, str):
      return where
    if isinstance(where, dict):
      return json.dumps(where, default=cls._json_encode)
    raise TypeError("expected json string or dict for parameter 'where'")


  @classmethod
  def _json_encode(cls, val):
    if isinstance(val, bytes):
      return "0x" + val.hex()
    if isinstance(val, datetime):
      return cls._datetime_to_ms(val)
    raise TypeError("expected only bytes or datetime")


  @classmethod
  def _timestamp_to_ms(cls, timestamp):
    if isinstance(timestamp, datetime):
      return cls._datetime_to_ms(timestamp)
    if not isinstance(timestamp, int):
      raise TypeError("couldn't parse {} as a timestamp".format(timestamp))
    return timestamp


  @classmethod
  def _datetime_to_ms(cls, dt):
    return int(dt.replace(tzinfo=timezone.utc).timestamp() * 1000)
