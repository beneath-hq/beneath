from collections.abc import Mapping
from datetime import timedelta
import logging
import os
from typing import Dict, Iterable, Union

from beneath import config
from beneath.admin.client import AdminClient
from beneath.checkpointer import Checkpointer, PrefixedCheckpointer
from beneath.config import DEFAULT_READ_BATCH_SIZE
from beneath.connection import Connection
from beneath.consumer import Consumer
from beneath.instance import StreamInstance
from beneath.job import Job
from beneath.stream import Stream
from beneath.utils import ProjectQualifier, StreamQualifier, SubscriptionQualifier
from beneath.writer import DryWriter, Writer


class Client:
    """
    The main class for interacting with Beneath.
    Data-related features (like defining streams and reading/writing data) are implemented
    directly on `Client`, while control-plane features (like creating projects) are isolated in
    the `admin` member.

    Args:
        secret (str):
            A beneath secret to use for authentication. If not set, uses the
            ``BENEATH_SECRET`` environment variable, and if that is not set either, uses the secret
            authenticated in the CLI (stored in ``~/.beneath``).
        dry (bool):
            If true, the client will not perform any mutations or writes, but generally perform
            reads as usual. It's useful for testing.

            The exact implication differs for different operations: Some mutations will be mocked,
            such as creating a stream, others will fail with an exception. Write operations log
            records to the logger instead of transmitting to the server. Reads generally work, but
            throw an exception when called on mocked resources.
        write_delay_ms (int):
            The maximum amount of time to buffer written records before sending
            a batch write request over the network. Defaults to 1 second (1000 ms).
            Writing records in batches reduces the number of requests, which leads to lower cost
            (Beneath charges at least 1kb per request).
    """

    def __init__(
        self,
        secret: str = None,
        dry: bool = False,
        write_delay_ms: int = config.DEFAULT_WRITE_DELAY_MS,
    ):
        self.connection = Connection(secret=self._get_secret(secret=secret))
        self.admin = AdminClient(connection=self.connection, dry=dry)
        self.dry = dry

        logging.basicConfig()
        self.logger = logging.getLogger("beneath")
        self.logger.setLevel(logging.INFO)

        if dry:
            self._writer = DryWriter(client=self, max_delay_ms=write_delay_ms)
        else:
            self._writer = Writer(client=self, max_delay_ms=write_delay_ms)

        self._start_count = 0
        self._checkpointers: Dict[ProjectQualifier, Checkpointer] = {}
        self._consumers: Dict[SubscriptionQualifier, Consumer] = {}

    @classmethod
    def _get_secret(cls, secret=None):
        if not secret:
            secret = os.getenv("BENEATH_SECRET", default=None)
        if not secret:
            secret = config.read_secret()
        if not secret:
            raise ValueError(
                "You must provide a secret (either authenticate with the CLI, set the "
                "BENEATH_SECRET environment variable, or pass a secret to the Client constructor)"
            )
        if not isinstance(secret, str):
            raise TypeError("secret must be a string")
        return secret.strip()

    # FINDING AND STAGING STREAMS

    async def find_stream(self, stream_path: str) -> Stream:
        """
        Finds an existing stream and returns an object that you can use to
        read and write from/to the stream.

        Args:
            path (str):
                The path to the stream in the format of "USERNAME/PROJECT/STREAM"
        """
        qualifier = StreamQualifier.from_path(stream_path)
        stream = await Stream._make(client=self, qualifier=qualifier)
        return stream

    async def create_stream(
        self,
        stream_path: str,
        schema: str,
        description: str = None,
        meta: bool = None,
        use_index: bool = None,
        use_warehouse: bool = None,
        retention: timedelta = None,
        log_retention: timedelta = None,
        index_retention: timedelta = None,
        warehouse_retention: timedelta = None,
        update_if_exists: bool = None,
    ) -> Stream:
        """
        Creates (or optionally updates if ``update_if_exists=True``) a stream and returns it.

        Args:
            stream_path (str):
                The (desired) path to the stream in the format of "USERNAME/PROJECT/STREAM".
                The project must already exist. If the stream doesn't exist yet, it creates it.
            schema (str):
                The GraphQL schema for the stream. To learn about the schema definition language,
                see https://about.beneath.dev/docs/reading-writing-data/schema-definition/.
            description (str):
                The description shown for the stream in the web console. If not set, tries to infer
                a description from the schema.
            retention (timedelta):
                The amount of time to retain records written to the stream.
                If not set, records will be stored forever.
            update_if_exists (bool):
                If true and the stream already exists, the provided info will be used to update
                the stream (only supports non-breaking schema changes) before returning it.
        """
        qualifier = StreamQualifier.from_path(stream_path)
        if self.dry:
            data = await self.admin.streams.compile_schema(
                schema_kind="GraphQL",
                schema=schema,
            )
            stream = await Stream._make_dry(
                client=self,
                qualifier=qualifier,
                avro_schema=data["canonicalAvroSchema"],
            )
        else:
            data = await self.admin.streams.create(
                organization_name=qualifier.organization,
                project_name=qualifier.project,
                stream_name=qualifier.stream,
                schema_kind="GraphQL",
                schema=schema,
                description=description,
                meta=meta,
                use_index=use_index,
                use_warehouse=use_warehouse,
                log_retention_seconds=self._timedelta_to_seconds(
                    log_retention if log_retention else retention
                ),
                index_retention_seconds=self._timedelta_to_seconds(
                    index_retention if index_retention else retention
                ),
                warehouse_retention_seconds=self._timedelta_to_seconds(
                    warehouse_retention if warehouse_retention else retention
                ),
                update_if_exists=update_if_exists,
            )
            stream = await Stream._make(client=self, qualifier=qualifier, admin_data=data)
        return stream

    @staticmethod
    def _timedelta_to_seconds(td: timedelta):
        if not td:
            return None
        secs = int(td.total_seconds())
        if secs == 0:
            return None
        return secs

    # WAREHOUSE QUERY

    async def query_warehouse(
        self,
        query: str,
        analyze: bool = False,
        max_bytes_scanned: int = config.DEFAULT_QUERY_WAREHOUSE_MAX_BYTES_SCANNED,
        timeout_ms: int = config.DEFAULT_QUERY_WAREHOUSE_TIMEOUT_MS,
    ):
        """
        Starts a warehouse (OLAP) SQL query, and returns a job for tracking its progress

        Args:
            query (str):
                The analytical SQL query to run. To learn about the query language,
                see https://about.beneath.dev/docs/reading-writing-data/warehouse-queries/.
            analyze (bool):
                If true, analyzes the query and returns info about referenced streams
                and expected bytes scanned, but doesn't actually run the query.
            max_bytes_scanned (int):
                Sets a limit on the number of bytes the query can scan.
                If exceeded, the job will fail with an error.
        """
        resp = await self.connection.query_warehouse(
            query=query,
            dry_run=analyze,
            max_bytes_scanned=max_bytes_scanned,
            timeout_ms=timeout_ms,
        )
        job_data = resp.job
        return Job(client=self, job_id=job_data.job_id, job_data=job_data)

    # WRITING

    async def start(self):
        """
        Opens the client for writes.
        Can be called multiple times, but make sure to call ``stop`` correspondingly.
        """
        self._start_count += 1
        if self._start_count != 1:
            return

        await self.connection.ensure_connected()
        await self._writer.start()

        for checkpointer in self._checkpointers.values():
            await checkpointer._start()

    async def stop(self):
        """
        Closes the client for writes, ensuring buffered writes are flushed.
        If ``start`` was called multiple times, only the last corresponding call
        to ``stop`` triggers a flush.
        """

        if self._start_count == 0:
            raise Exception("Called stop more times than start")

        if self._start_count == 1:
            for checkpointer in self._checkpointers.values():
                await checkpointer._stop()
            await self._writer.stop()

        self._start_count -= 1

    async def write(self, instance: StreamInstance, records: Union[Mapping, Iterable[Mapping]]):
        """
        Writes one or more records to ``instance``. By default, writes are buffered for up to
        ``write_delay_ms`` milliseconds before being transmitted to the server. See the Client
        constructor for details.

        To enabled writes, make sure to call ``start`` on the client (and ``stop`` before
        terminating).

        Args:
            instance (StreamInstance):
                The instance to write to. You can also call ``instance.write`` as a convenience
                wrapper.
            records:
                The records to write. Can be a single record (dict) or a list of records (iterable
                of dict).
        """
        if self._start_count == 0:
            raise Exception("Cannot call write because the client is stopped")
        await self._writer.write(instance, records)

    async def force_flush(self):
        """ Forces the client to flush buffered writes without stopping """
        await self._writer.force_flush()

    # CHECKPOINTERS

    async def checkpointer(
        self,
        project_path: str,
        metastream_name="checkpoints",
        key_prefix: str = None,
        create_metastream: bool = True,
    ) -> Checkpointer:
        """
        Returns a checkpointer for the given project.
        Checkpointers store (small) key-value records useful for maintaining consumer and pipeline
        state. State is stored in a meta-stream called "checkpoints" in the given project.

        Args:
            project_path (str):
                Path to the project in which to store the checkpointer's state
            metastream_name (str):
                Name of the meta stream in which to save checkpointed data
            create_metastream (bool):
                If true, the checkpointer will create the "checkpoints" meta-stream if it does not
                already exists. If false, the checkpointer will throw an exception if the
                meta-stream does not already exist. Defaults to True.
            key_prefix (str):
                If set, any ``get`` or ``set`` call on the checkpointer will prepend the prefix
                to the key.
        """
        project_qualifier = ProjectQualifier.from_path(project_path)
        qualifier = StreamQualifier(
            organization=project_qualifier.organization,
            project=project_qualifier.project,
            stream=metastream_name,
        )

        if qualifier not in self._checkpointers:
            checkpointer = Checkpointer(
                self,
                qualifier,
                create_metastream=create_metastream,
            )
            self._checkpointers[qualifier] = checkpointer
            if self._start_count != 0:
                await checkpointer._start()

        checkpointer = self._checkpointers[qualifier]
        if key_prefix:
            checkpointer = PrefixedCheckpointer(checkpointer, key_prefix)
        return checkpointer

    # CONSUMERS

    async def consumer(
        self,
        stream_path: str,
        version: int = None,
        batch_size: int = DEFAULT_READ_BATCH_SIZE,
        subscription_path: str = None,
        checkpointer: Checkpointer = None,
        create_metastream: bool = True,
    ):
        """
        Creates a consumer for the given stream.
        Consumers make it easy to replay the history of a stream and/or subscribe to new changes.

        Args:
            stream_path (str):
                Path to the stream to subscribe to. The consumer will subscribe to the stream's
                primary version.
            version (int):
                The instance version to use for stream. If not set, uses the primary instance.
            batch_size (int):
                Sets the max number of records to load in each network request. Defaults to 1000.
            subscription_path (str):
                Format "ORGANIZATION/PROJECT/NAME". If set, the consumer will use a checkpointer
                to save cursors. That means processing will not restart from scratch if the process
                ends or crashes (as long as you use the same subscription name). To reset a
                subscription, call ``reset`` on the consumer.
            checkpointer (Checkpointer):
                Only applies if ``subscription_path`` is set. Provides a specific checkpointer to
                use for consumer state. If not set, will create one in the subscription's project.
            create_metastream (bool):
                Only applies if ``subscription_path`` is set and ``checkpointer`` is not set.
                Passed through to ``client.checkpointer``.
        """
        stream_qualifier = StreamQualifier.from_path(stream_path)
        if not subscription_path:
            consumer = Consumer(
                client=self,
                stream_qualifier=stream_qualifier,
                batch_size=batch_size,
            )
            await consumer._init()
            return consumer

        sub_qualifier = SubscriptionQualifier.from_path(subscription_path)
        if sub_qualifier not in self._consumers:
            if checkpointer is None:
                checkpointer = await self.checkpointer(
                    f"{sub_qualifier.organization}/{sub_qualifier.project}",
                    create_metastream=create_metastream,
                )
            consumer = Consumer(
                client=self,
                stream_qualifier=stream_qualifier,
                version=version,
                batch_size=batch_size,
                checkpointer=checkpointer,
                subscription_name=sub_qualifier.subscription,
            )
            await consumer._init()
            self._consumers[sub_qualifier] = consumer

        return self._consumers[sub_qualifier]
