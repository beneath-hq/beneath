# allows us to use Client as a type hint without an import cycle
# see: https://www.stefaanlippens.net/circular-imports-type-hints-python.html
# pylint: disable=wrong-import-position,ungrouped-imports
from __future__ import annotations
from typing import TYPE_CHECKING

if TYPE_CHECKING:
    from beneath.client import Client

from datetime import timedelta
import json
import sys
from typing import Any, Dict, Tuple

import msgpack

from beneath.config import DEFAULT_CHECKPOINT_COMMIT_DELAY_MS
from beneath.instance import StreamInstance
from beneath.utils import AIODelayBuffer, StreamQualifier

SERVICE_CHECKPOINT_LOG_RETENTION = timedelta(hours=6)
SERVICE_CHECKPOINT_SCHEMA = """
  type Checkpoint @schema {
    key: String! @key
    value: Bytes
  }
"""


class Checkpointer:
    """
    Checkpointers store (small) key-value records in a meta stream (in the specified project).
    They are useful for maintaining consumer and pipeline state, such as the cursor for a
    subscription or the last time a scraper ran.

    Checkpoint keys are strings and values are serialized with msgpack (supports most normal Python
    values, but not custom classes).

    New checkpointed values are flushed at regular intervals (every 30 seconds by default).
    Checkpointers should not be used for storing large amounts of data.
    Checkpointers are not currently suitable for synchronizing parallel processes.
    """

    instance: StreamInstance
    """ The meta-stream instance that checkpoints are written to """

    def __init__(
        self,
        client: Client,
        metastream_qualifier: StreamQualifier,
        metastream_create: bool,
        metastream_description: str,
    ):
        self.instance = None
        self._client = client
        self._metastream_qualifier = metastream_qualifier
        self._metastream_description = metastream_description
        self._create = metastream_create
        self._writer = self._Writer(self)
        self._cache: Dict[str, Any] = {}

    async def get(self, key: str, default: Any = None) -> Any:
        """
        Gets a checkpointed value
        """
        if key in self._cache:
            return self._cache[key]

        if self._client.dry:
            return default

        filt = json.dumps({"key": key})
        cursor = await self.instance.query_index(filter=filt)

        value = default
        record = await cursor.read_one()
        if record is not None:
            value = msgpack.unpackb(record["value"], timestamp=3)

        # checking self._cache again because of awaits (and we'd rather serve a recent local set)
        if key not in self._cache:
            self._cache[key] = value

        return self._cache[key]

    async def set(self, key: str, value: Any):
        """ Sets a checkpoint value. Value will be encoded with msgpack. """
        if not self._writer.running:
            raise Exception("Cannot call 'set' on checkpointer because the client is stopped")
        self._cache[key] = value
        await self._writer.write(key, value)

    # START/STOP (called by client)

    async def _start(self):
        if not self.instance:
            await self._stage_stream()
        await self._writer.start()

    async def _stop(self):
        await self._writer.stop()

    # CHECKPOINT STREAM

    async def _stage_stream(self):
        if self._create or self._client.dry:
            stream = await self._client.create_stream(
                stream_path=str(self._metastream_qualifier),
                schema=SERVICE_CHECKPOINT_SCHEMA,
                description=self._metastream_description,
                meta=True,
                log_retention=SERVICE_CHECKPOINT_LOG_RETENTION,
                use_warehouse=False,
                update_if_exists=True,
            )
        else:
            stream = await self._client.find_stream(stream_path=str(self._metastream_qualifier))

        if not stream.primary_instance:
            raise Exception("Expected checkpoints stream to have a primary instance")
        self.instance = stream.primary_instance

        self._client.logger.info(
            "Using '%s' (version %i) for checkpointing",
            self._metastream_qualifier,
            self.instance.version,
        )

    # CHECKPOINT WRITER

    class _Writer(AIODelayBuffer[Tuple[str, Any]]):

        checkpoint: Dict[str, Any]

        def __init__(self, checkpointer: Checkpointer):
            super().__init__(
                max_delay_ms=DEFAULT_CHECKPOINT_COMMIT_DELAY_MS,
                max_record_size=sys.maxsize,
                max_buffer_size=sys.maxsize,
                max_buffer_count=sys.maxsize,
            )
            self.checkpointer = checkpointer

        def _reset(self):
            self.checkpoint = {}

        def _merge(self, value: Tuple[str, Any]):
            (key, checkpoint) = value
            self.checkpoint[key] = checkpoint

        async def _flush(self):
            records = (
                {
                    "key": key,
                    "value": msgpack.packb(checkpoint, datetime=True),
                }
                for (key, checkpoint) in self.checkpoint.items()
            )
            if self.checkpointer.instance:
                await self.checkpointer._client.write(
                    instance=self.checkpointer.instance,
                    records=records,
                )

        # pylint: disable=arguments-differ
        async def write(self, key: str, checkpoint: Any):
            await super().write(value=(key, checkpoint), size=0)


class PrefixedCheckpointer:
    """ Wraps a Checkpointer and prefixes all keys """

    def __init__(self, checkpointer: Checkpointer, prefix: str):
        self._checkpointer = checkpointer
        self._prefix = prefix

    @property
    def instance(self):
        return self._checkpointer.instance

    async def get(self, key: str, default: Any = None) -> Any:
        key = self._prefix + key
        return await self._checkpointer.get(key, default=default)

    async def set(self, key: str, value: Any):
        key = self._prefix + key
        return await self._checkpointer.set(key, value)
