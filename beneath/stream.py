import io
from datetime import datetime
import json
import sys
import warnings

from fastavro import parse_schema
from fastavro import schemaless_reader
from fastavro import schemaless_writer
import pandas as pd

from beneath import config
from beneath.proto import engine_pb2
from beneath.utils import datetime_to_ms
from beneath.utils import timestamp_to_ms
from beneath.utils import ms_to_datetime
from beneath.utils import ms_to_pd_timestamp

class Stream:
  """
  Stream enables read and write operations on Beneath streams
  """

  def __init__(self, client, stream_id, project_name, stream_name, schema, key_fields, avro_schema, batch, current_instance_id):
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
    self.stream_id = stream_id
    self.project_name = project_name
    self.stream_name = stream_name
    self.schema = schema
    self.key_fields = key_fields
    self.avro_schema = avro_schema
    self.parsed_avro_schema = parse_schema(json.loads(avro_schema))
    self.batch = batch
    self.current_instance_id = current_instance_id


  def __getstate__(self):
    return {
      'client': self.client,
      'stream_id': self.stream_id,
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
    self.stream_id = obj['stream_id']
    self.project_name = obj['project_name']
    self.stream_name = obj['stream_name']
    self.schema = obj['schema']
    self.key_fields = obj['key_fields']
    self.avro_schema = obj['avro_schema']
    self.parsed_avro_schema = parse_schema(json.loads(self.avro_schema))
    self.batch = obj['batch']
    self.current_instance_id = obj['current_instance_id']


  def write(self, records, instance_id=None):
    instance_id = self._instance_id_or_default(instance_id)
    encoded_records = (self._encode_record(record) for record in records)

    batch = []
    batch_bytes = 0
    def flush():
      nonlocal batch
      nonlocal batch_bytes
      if len(batch) > 0:
        self.client.write_batch(instance_id, batch)
        batch = []
        batch_bytes = 0
    
    for er in encoded_records:
      if len(batch) == config.WRITE_BATCH_SIZE:
        flush()
      if batch_bytes + len(er.avro_data) > config.WRITE_BATCH_BYTES:
        flush()
      batch.append(er)
      batch_bytes += len(er.avro_data)

    flush()


  def read(self, limit=None, where=None, to_dataframe=True, max_rows=None, max_megabytes=None, instance_id=None, warn_max=True):
    instance_id = self._instance_id_or_default(instance_id)
    where = self._parse_where(where)

    def read(limit, last):
      after = None
      if last:
        after = {field: last[field] for field in self.key_fields}
        after = json.dumps(after, default=self._json_encode)
      return self.client.read_batch(instance_id=instance_id, where=where, limit=limit, after=after)

    return self._constrained_read(
      read_fn=read,
      limit=limit,
      to_dataframe=to_dataframe,
      max_rows=max_rows,
      max_megabytes=max_megabytes,
      warn_max=warn_max,
    )


  def latest(self, limit=None, to_dataframe=True, max_rows=None, max_megabytes=None, instance_id=None, warn_max=True):
    instance_id = self._instance_id_or_default(instance_id)

    def read(limit, last):
      before = None
      if last:
        before = last["@meta.timestamp"]
        if to_dataframe:
          before = before.to_pydatetime()
      return self.client.read_latest_batch(instance_id=instance_id, limit=limit, before=before)

    return self._constrained_read(
      read_fn=read,
      limit=limit,
      to_dataframe=to_dataframe,
      max_rows=max_rows,
      max_megabytes=max_megabytes,
      warn_max=warn_max,
    )


  def search_range_end(self, target, end_max, where=None):
    # defaults
    where = where if where else {}

    # validate fields
    if target not in self.key_fields:
      raise ValueError(f"target field ('{target}') is not a key field")
    if target in where:
      raise ValueError(f"cannot search target field ('{target}') when it is specified in the where arg")
    for f in self.key_fields:
      if f == target:
        break
      if f not in where:
        raise ValueError(f"for fields that are to the left of the target field ('{target}') in the stream key, you must set the where arg")

    # the function we use to divide 
    middle = None
    if self._column_is_integral(target):
      middle = lambda x1, x2: x1 + int((x2 - x1) / 2)
    elif self._column_is_timestamp(target):
      middle = lambda t1, t2: ms_to_datetime(datetime_to_ms(t1 + ((t2 - t1) / 2)))

    # invariant: we know it's between these
    start = None
    end = end_max

    # check there exists a maximum (and handle the easy case where there's only one row)
    where[target] = { "_lt": end }
    rows = self.read(where=where, limit=2, to_dataframe=False)
    if len(rows) == 0:
      return None
    elif len(rows) == 1:
      return rows[0][target]
    else:
      start = rows[-1][target]

    # search
    iterations = 0
    while iterations < 64:
      mid = middle(start, end)
      if start == mid: 
        # for non-complete keys, can be multiple rows with max, which would cause infinite loop
        # note: == is well-defined for datetime
        return start

      where[target] = { "_gte": mid, "_lt": end }
      rows = self.read(where=where, limit=2, to_dataframe=False)
      
      if len(rows) == 0: # too far
        end = mid
      elif len(rows) == 1: # found it
        return rows[0][target]
      else: # not far enough
        start = rows[-1][target]
      
      iterations += 1
    
    raise RuntimeError(f"stuck in infinite search at start={start} end={end}")


  def _constrained_read(self, read_fn, limit=None, to_dataframe=True, max_rows=None, max_megabytes=None, warn_max=True):
    # adding 1 to turn <= into <
    max_rows = min(x for x in [limit, max_rows, sys.maxsize] if x)
    max_megabytes = max_megabytes if max_megabytes else config.DEFAULT_MAX_READ_MB
    max_bytes = self._mb_to_bytes(max_megabytes) + 1

    records = []
    complete = False
    rows_loaded = 0
    bytes_loaded = 0

    while rows_loaded < max_rows and bytes_loaded < max_bytes:
      last = None
      if len(records) > 0:
        last = records[-1]

      lim = min(max_rows - rows_loaded, config.READ_BATCH_SIZE)
      if lim == 0:
        break

      batch = read_fn(limit=lim, last=last)
      if len(batch) == 0:
        complete = True
        break

      for record in batch:
        obj = self._decode_avro(record.avro_data)
        obj["@meta.timestamp"] = ms_to_pd_timestamp(record.timestamp) if to_dataframe else ms_to_datetime(record.timestamp)
        records.append(obj)
        rows_loaded += 1
        bytes_loaded += len(record.avro_data)
        if bytes_loaded >= max_bytes:
          break

      if len(batch) < lim:
        complete = True
        break

      if limit and rows_loaded == limit:
        # fundamentally the difference between limit and max_rows
        complete = True

    if not complete and warn_max:
      # Jupyter doesn't always display warnings, so also print
      if rows_loaded >= max_rows:
        err = "Stopped loading because stream length exceeds max_rows={}".format(max_rows)
        print(err)
        warnings.warn(err)
      elif bytes_loaded >= max_bytes:
        err = "Stopped loading because download size exceeds max_megabytes={}MB".format(max_megabytes)
        print(err)
        warnings.warn(err)

    if to_dataframe:
      return pd.DataFrame(records, columns=self._columns)
    return records


  @property
  def _columns(self):
    return [field["name"] for field in self.parsed_avro_schema["fields"]]


  @property
  def bigquery_table(self):
    return self.get_bigquery_table()


  def _column_is_integral(self, name):
    for field in self.parsed_avro_schema["fields"]:
      if field["name"] == name:
        return field["type"] == "int" or field["type"] == "long"
    raise ValueError(f"field '{name}' not found")


  def _column_is_timestamp(self, name):
    for field in self.parsed_avro_schema["fields"]:
      if field["name"] == name:
        if isinstance(field["type"], dict):
          return field["type"].get("logicalType", None) == "timestamp-millis"
        return False
    raise ValueError(f"field '{name}' not found")


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
      return timestamp_to_ms(record["@meta"]["timestamp"])
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
    record = schemaless_reader(reader, self.parsed_avro_schema)
    reader.close()
    return record


  def _encode_avro(self, record):
    writer = io.BytesIO()
    schemaless_writer(writer, self.parsed_avro_schema, record)
    result = writer.getvalue()
    writer.close()
    return result


  @classmethod
  def _mb_to_bytes(cls, mbs):
    return mbs * (2**20)


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
      return datetime_to_ms(val)
    raise TypeError("expected only bytes or datetime")
