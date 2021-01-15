from datetime import timedelta
import os

from beneath import config
from beneath.admin.client import AdminClient
from beneath.connection import Connection
from beneath.job import Job
from beneath.stream import Stream
from beneath.utils import StreamQualifier
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
    """

    def __init__(self, secret=None):
        self.connection = Connection(secret=self._get_secret(secret=secret))
        self.admin = AdminClient(connection=self.connection)

    @classmethod
    def _get_secret(cls, secret=None):
        if not secret:
            secret = os.getenv("BENEATH_SECRET", default=None)
        if not secret:
            secret = config.read_secret()
        if not secret:
            raise ValueError(
                "you must provide a secret (either authenticate with the CLI, set the "
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
        stream = await Stream.make(client=self, qualifier=qualifier)
        return stream

    async def create_stream(
        self,
        stream_path: str,
        schema: str,
        description: str = None,
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
            retention (timedelta):
                The amount of time to retain records written to the stream.
                If not set, records will be stored forever.
            update_if_exists (bool):
                If true and the stream already exists, the provided info will be used to update
                the stream (only supports non-breaking schema changes) before returning it.
        """
        qualifier = StreamQualifier.from_path(stream_path)
        data = await self.admin.streams.create(
            organization_name=qualifier.organization,
            project_name=qualifier.project,
            stream_name=qualifier.stream,
            schema_kind="GraphQL",
            schema=schema,
            description=description,
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
        stream = await Stream.make(client=self, qualifier=qualifier, admin_data=data)
        return stream

    @staticmethod
    def _timedelta_to_seconds(td: timedelta):
        if not td:
            return None
        secs = int(td.total_seconds())
        if secs == 0:
            return None
        return secs

    # WRITING

    def writer(self, dry=False, write_delay_ms: int = config.DEFAULT_WRITE_DELAY_MS) -> Writer:
        """
        Return a ``Writer`` object for writing data to multiple streams at once.
        A ``Writer`` buffers records in memory for up to 1 second (by default) before
        sending them in batches over the network.

        Args:
            dry (bool):
                If true, written records will be printed, not transmitted to the server.
                Useful for testing.
            write_delay_ms (int):
                The maximum amount of time to buffer records before sending a
                batch write over the network. Defaults to 1 second (1000 ms).
        """
        if dry:
            return DryWriter(max_delay_ms=write_delay_ms)
        return Writer(connection=self.connection, max_delay_ms=write_delay_ms)

    # WAREHOUSE QUERY

    async def query_warehouse(
        self,
        query: str,
        dry: bool = False,
        max_bytes_scanned: int = config.DEFAULT_QUERY_WAREHOUSE_MAX_BYTES_SCANNED,
        timeout_ms: int = config.DEFAULT_QUERY_WAREHOUSE_TIMEOUT_MS,
    ):
        """
        Starts a warehouse (OLAP) SQL query, and returns a job for tracking its progress

        Args:
            query (str):
                The analytical SQL query to run. To learn about the query language,
                see https://about.beneath.dev/docs/reading-writing-data/warehouse-queries/.
            dry (bool):
                If true, analyzes the query and returns info about referenced streams
                and expected bytes scanned, but doesn't actually run the query.
            max_bytes_scanned (int):
                Sets a limit on the number of bytes the query can scan.
                If exceeded, the job will fail with an error.
        """
        resp = await self.connection.query_warehouse(
            query=query,
            dry_run=dry,
            max_bytes_scanned=max_bytes_scanned,
            timeout_ms=timeout_ms,
        )
        job_data = resp.job
        return Job(client=self, job_id=job_data.job_id, job_data=job_data)
