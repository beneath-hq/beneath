from collections.abc import Mapping
import json
import io
from typing import Tuple

from fastavro import parse_schema
from fastavro import schemaless_reader
from fastavro import schemaless_writer

from beneath.proto import gateway_pb2
from beneath.utils import (
    ms_to_datetime,
    ms_to_pd_timestamp,
    timestamp_to_ms,
)


class Schema:
    """ Represents a stream's parsed Avro schema """

    def __init__(self, avro: str):
        self.parsed_avro = parse_schema(json.loads(avro))
        """ The parsed avro schema """

    def record_to_pb(self, record: Mapping) -> Tuple[gateway_pb2.Record, int]:
        if not isinstance(record, Mapping):
            raise TypeError("write error: record must be a mapping, got {}".format(record))
        avro = self._encode_avro(record)
        timestamp = self._extract_record_timestamp(record)
        pb = gateway_pb2.Record(avro_data=avro, timestamp=timestamp)
        return (pb, pb.ByteSize())

    def pb_to_record(self, pb: gateway_pb2.Record, to_dataframe: bool) -> Mapping:
        record = self._decode_avro(pb.avro_data)
        record["@meta.timestamp"] = (
            ms_to_pd_timestamp(pb.timestamp) if to_dataframe else ms_to_datetime(pb.timestamp)
        )
        return record

    def _encode_avro(self, record: Mapping):
        writer = io.BytesIO()
        schemaless_writer(writer, self.parsed_avro, record)
        result = writer.getvalue()
        writer.close()
        return result

    def _decode_avro(self, data):
        reader = io.BytesIO(data)
        record = schemaless_reader(reader, self.parsed_avro)
        reader.close()
        return record

    @staticmethod
    def _extract_record_timestamp(record: Mapping) -> int:
        if ("@meta" in record) and ("timestamp" in record["@meta"]):
            return timestamp_to_ms(record["@meta"]["timestamp"])
        return 0  # 0 tells the server to set timestamp to its current time
