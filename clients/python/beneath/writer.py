# Allows us to use Stream as a type hint without an import cycle
# pylint: disable=wrong-import-position,ungrouped-imports
from __future__ import annotations
from typing import TYPE_CHECKING

if TYPE_CHECKING:
    from beneath.client import Client
    from beneath.instance import StreamInstance

from collections import defaultdict
from collections.abc import Mapping
from typing import Dict, Iterable, List, Union, Tuple
import uuid

from beneath import config
from beneath.proto import gateway_pb2
from beneath.utils import AIODelayBuffer

InstanceIDAndRecordPB = Tuple[uuid.UUID, gateway_pb2.Record]
InstanceRecordAndSize = Tuple["StreamInstance", Mapping, int]


class Writer(AIODelayBuffer[InstanceIDAndRecordPB]):
    """ Override of AIODelayBuffer designed to buffer and write to multiple instances at once """

    _client: Client
    _records: Dict[uuid.UUID, List[gateway_pb2.Record]]
    _total: int

    def __init__(self, client: Client, max_delay_ms: int):
        super().__init__(
            max_delay_ms=max_delay_ms,
            max_record_size=config.MAX_RECORD_SIZE_BYTES,
            max_buffer_size=config.MAX_BATCH_SIZE_BYTES,
            max_buffer_count=config.MAX_BATCH_SIZE_COUNT,
        )
        self._client = client
        self._total = 0

    def _reset(self):
        self._records = defaultdict(list)

    def _merge(self, value: InstanceIDAndRecordPB):
        (instance_id, record) = value
        self._records[instance_id].append(record)

    async def _flush(self):
        await self._client.connection.write(
            [
                gateway_pb2.InstanceRecords(instance_id=instance_id.bytes, records=record_pbs)
                for (instance_id, record_pbs) in self._records.items()
            ]
        )
        count = 0
        for (_, record_pbs) in self._records.items():
            count += len(record_pbs)
        self._total += count
        self._client.logger.info(
            "Flushed %i records to %i instances (%i total during session)",
            count,
            len(self._records),
            self._total,
        )

    # pylint: disable=arguments-differ
    async def write(self, instance: StreamInstance, records: Union[Mapping, Iterable[Mapping]]):
        if isinstance(records, Mapping):
            records = [records]
        for record in records:
            (pb, size) = instance.stream.schema.record_to_pb(record)
            await super().write(value=(instance.instance_id, pb), size=size)


class DryWriter(AIODelayBuffer[InstanceRecordAndSize]):
    """ Override of AIODelayBuffer designed to buffer and write to multiple instances at once """

    _client: Client
    _records: List[InstanceRecordAndSize]
    _total: int

    def __init__(self, client: Client, max_delay_ms: int):
        super().__init__(
            max_delay_ms=max_delay_ms,
            max_record_size=config.MAX_RECORD_SIZE_BYTES,
            max_buffer_size=config.MAX_BATCH_SIZE_BYTES,
            max_buffer_count=config.MAX_BATCH_SIZE_COUNT,
        )
        self._client = client
        self._total = 0

    def _reset(self):
        self._records = []

    def _merge(self, value: InstanceRecordAndSize):
        self._records.append(value)

    async def _flush(self):
        self._client.logger.info("Flushing %i buffered records", len(self._records))
        for value in self._records:
            (instance, record, size) = value
            self._client.logger.info(
                "Flushed record (stream=%s, size=%i bytes): %s",
                str(instance.stream._qualifier),
                size,
                record,
            )
            self._total += 1
        self._client.logger.info(
            "Flushed %i records (%i total during session)",
            len(self._records),
            self._total,
        )

    # pylint: disable=arguments-differ
    async def write(self, instance: StreamInstance, records: Union[Mapping, Iterable[Mapping]]):
        if isinstance(records, Mapping):
            records = [records]
        for record in records:
            (_, size) = instance.stream.schema.record_to_pb(record)
            value: InstanceRecordAndSize = (instance, record, size)
            await super().write(value=value, size=size)
