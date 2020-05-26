from datetime import timedelta
from collections.abc import Mapping
import os
from typing import Awaitable, Callable, Iterable

from beneath import __version__
from beneath import config
from beneath.stream import Stream
from beneath.admin.models import Models
from beneath.admin.organizations import Organizations
from beneath.admin.projects import Projects
from beneath.admin.secrets import Secrets
from beneath.admin.services import Services
from beneath.admin.streams import Streams
from beneath.admin.users import Users
from beneath.connection import Connection, GraphQLError
from beneath.config import (
  DEFAULT_READ_ALL_MAX_BYTES,
  DEFAULT_READ_BATCH_SIZE,
  DEFAULT_SUBSCRIBE_CONCURRENT_CALLBACKS,
  DEFAULT_SUBSCRIBE_PREFETCHED_RECORDS,
)
from beneath.utils import StreamQualifier


class Client:
  """
  The main class for interacting with Beneath.
  Data-related features (like defining streams and reading/writing data) are implemented
  directly on `Client`, while control-plane features (like creating projects) are isolated in
  the `admin` member.

  Kwargs:
    secret (str): A beneath secret to use for authentication. If not set, reads secret from ``~/.beneath``.
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
    if not isinstance(secret, str):
      raise TypeError("secret must be a string")
    return secret.strip()

  async def find_stream(self, path: str) -> Stream:
    """
    Finds an existing stream and returns an object that you can use to 
    read and write from/to the stream.

    Args:
      path (str): The path to the stream in the format of "ORGANIZATION/PROJECT/STREAM"
    """
    qualifier = StreamQualifier.from_path(path)
    stream = Stream(client=self, qualifier=qualifier)
    # pylint: disable=protected-access
    await stream._ensure_loaded()
    return stream

  async def stage_stream(
    self,
    path: str,
    schema: str,
    retention: timedelta = None,
    create_primary_instance: bool = True,
  ) -> Stream:
    """
    The one-stop call for creating, updating and getting a stream:
    a) If the stream doesn't exist, it creates it, then returns it.
    b) If the stream exists and you have changed the schema, it updates the stream's schema (only supports non-breaking changes), then returns it.
    c) If the stream exists and the schema matches, it fetches the stream and returns it.

    Args:
      path (str): The (desired) path to the stream in the format of "ORGANIZATION/PROJECT/STREAM".
        The project must already exist. If the stream doesn't exist yet, it creates it.
      schema (str): The GraphQL schema for the stream.
        To learn about the schema definition language, see https://about.beneath.dev/docs/reading-writing-data/creating-streams/).
    
    Kwargs:
      retention (timedelta): The amount of time to retain records written to the stream.
        If not set, records will be stored forever.
      create_primary_instance (bool): If true (default), automatically creates a primary instance for the stream.
        To learn more about stream instances, see https://about.beneath.dev/docs/reading-writing-data/understanding-streams/.
    """
    qualifier = StreamQualifier.from_path(path)
    data = await self.admin.streams.stage(
      organization_name=qualifier.organization,
      project_name=qualifier.project,
      stream_name=qualifier.stream,
      schema_kind="GraphQL",
      schema=schema,
      retention_seconds=retention.seconds if retention else None,
      create_primary_instance=create_primary_instance,
    )
    stream = Stream(client=self, qualifier=qualifier)
    # pylint: disable=protected-access
    await stream._ensure_loaded(prefetched=data)
    return stream

  # EASY HELPERS

  async def easy_read(
    self,
    stream_path: str,
    # pylint: disable=redefined-builtin
    filter: str = None,
    to_dataframe=True,
    batch_size=DEFAULT_READ_BATCH_SIZE,
    max_bytes=DEFAULT_READ_ALL_MAX_BYTES,
    warn_max=True,
  ) -> Iterable[Mapping]:
    """
    Helper function for reading all records in a stream into memory (as a list or as a `pandas.DataFrame`).

    Args:
      stream_path (str): Path to the stream to read from, format: "ORGANIZATION/PROJECT/STREAM".

    Kwargs:
      filter (str): Optional filter to only retrieve specific records.
        For details about the filter syntax, see https://about.beneath.dev/docs/reading-writing-data/indexed-reads/.
      
      to_dataframe (bool): If true, returns the results as a Pandas DataFrame, otherwise returns a list.
      
      batch_size (int): The number of records to fetch per request (`easy_read` issues many requests in order to load the entire stream).
      
      max_bytes (int): Limits the amount of (serialized) data loaded. Causes `easy_read` to return a partial result if the download exceeds `max_bytes`.
      
      warn_max (int): Whether to issue a warning if `max_bytes` is reached.
    """
    stream = await self.find_stream(path=stream_path)
    cursor = await stream.query_index(filter=filter)
    res = await cursor.fetch_all(
      max_bytes=max_bytes,
      batch_size=batch_size,
      warn_max=warn_max,
      to_dataframe=to_dataframe,
    )
    return res

  async def easy_process_once(
    self,
    stream_path: str,
    callback: Callable[[Mapping], Awaitable[None]],
    # pylint: disable=redefined-builtin
    filter: str = None,
    max_prefetched_records=DEFAULT_SUBSCRIBE_PREFETCHED_RECORDS,
    max_concurrent_callbacks=DEFAULT_SUBSCRIBE_CONCURRENT_CALLBACKS,
  ):
    """
    Helper function for calling a function on every record in a stream.
    It may process multiple records at once (see `max_concurrent_callbacks`).
    It can be used on large streams as it does *not* load the entire stream into memory at once.

    Args:
      stream_path (str): Path to the stream to read from, format: "ORGANIZATION/PROJECT/STREAM".

      callback (async fn): Function to call on every record.
    
    Kwargs:
      filter (str): Optional filter to only retrieve specific records.
        For details about the filter syntax, see https://about.beneath.dev/docs/reading-writing-data/indexed-reads/.
      
      max_prefetched_records (int): Max number of records to prefetch (buffer) before processing.

      max_concurrent_callbacks (int): Max number of callbacks to run in parallel.
        When it reaches this limit, it waits for one of the outstanding callbacks to finish before triggering a new callback.
    """
    stream = await self.find_stream(path=stream_path)
    cursor = await stream.query_index(filter=filter)
    await cursor.subscribe_replay(
      callback=callback,
      max_prefetched_records=max_prefetched_records,
      max_concurrent_callbacks=max_concurrent_callbacks,
    )

  async def easy_process_forever(
    self,
    stream_path: str,
    callback: Callable[[Mapping], Awaitable[None]],
    max_prefetched_records=DEFAULT_SUBSCRIBE_PREFETCHED_RECORDS,
    max_concurrent_callbacks=DEFAULT_SUBSCRIBE_CONCURRENT_CALLBACKS,
  ):
    """
    Helper function to replay every past record in a stream, and then subscribe to new incoming changes.
    Calls `callback` on every record fetched.
    It may process multiple records at once (see `max_concurrent_callbacks`).
    It can be used on large streams as it does *not* load the entire stream into memory at once.

    Args:
      stream_path (str): Path to the stream to read from, format: "ORGANIZATION/PROJECT/STREAM".

      callback (async fn): Function to call on every record.
    
    Kwargs:
      max_prefetched_records (int): Max number of records to prefetch (buffer) before processing.

      max_concurrent_callbacks (int): Max number of callbacks to run in parallel.
        When it reaches this limit, it waits for one of the outstanding callbacks to finish before triggering a new callback.
    """
    stream = await self.find_stream(path=stream_path)
    cursor = await stream.query_index()
    await cursor.subscribe_replay(
      callback=callback,
      max_prefetched_records=max_prefetched_records,
      max_concurrent_callbacks=max_concurrent_callbacks,
    )
    await cursor.subscribe_changes(
      callback=callback,
      max_prefetched_records=max_prefetched_records,
      max_concurrent_callbacks=max_concurrent_callbacks,
    )


class AdminClient:
  """
  AdminClient isolates control-plane features.

  Args:
      connection (Connection): An authenticated connection to Beneath.
  """

  def __init__(self, connection: Connection):
    self.connection = connection
    self.models = Models(self.connection)
    self.organizations = Organizations(self.connection)
    self.projects = Projects(self.connection)
    self.secrets = Secrets(self.connection)
    self.services = Services(self.connection)
    self.streams = Streams(self.connection)
    self.users = Users(self.connection)
