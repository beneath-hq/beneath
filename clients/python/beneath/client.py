from datetime import timedelta
import os

from beneath import __version__
from beneath import config
from beneath.admin.client import AdminClient
from beneath.connection import Connection
from beneath.stream import Stream
from beneath.utils import StreamQualifier
from beneath.writer import DryWriter, Writer


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

  # FINDING AND STAGING STREAMS

  async def find_stream(self, stream_path: str) -> Stream:
    """
    Finds an existing stream and returns an object that you can use to
    read and write from/to the stream.

    Args:
      path (str): The path to the stream in the format of "ORGANIZATION/PROJECT/STREAM"
    """
    qualifier = StreamQualifier.from_path(stream_path)
    stream = await Stream.make(client=self, qualifier=qualifier)
    return stream

  async def stage_stream(
    self,
    stream_path: str,
    schema: str,
    use_index: bool = None,
    use_warehouse: bool = None,
    log_retention: timedelta = None,
    index_retention: timedelta = None,
    warehouse_retention: timedelta = None,
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
    """
    qualifier = StreamQualifier.from_path(stream_path)
    data = await self.admin.streams.stage(
      organization_name=qualifier.organization,
      project_name=qualifier.project,
      stream_name=qualifier.stream,
      schema_kind="GraphQL",
      schema=schema,
      use_index=use_index,
      use_warehouse=use_warehouse,
      log_retention_seconds=int(log_retention.total_seconds()) if log_retention else None,
      index_retention_seconds=int(index_retention.total_seconds()) if index_retention else None,
      warehouse_retention_seconds=int(warehouse_retention.total_seconds()) if warehouse_retention else None,
    )
    stream = await Stream.make(client=self, qualifier=qualifier, admin_data=data)
    return stream

  # WRITING

  def writer(self, dry=False, write_delay_ms: int = config.DEFAULT_WRITE_DELAY_MS) -> Writer:
    if dry:
      return DryWriter(max_delay_ms=write_delay_ms)
    return Writer(connection=self.connection, max_delay_ms=write_delay_ms)
