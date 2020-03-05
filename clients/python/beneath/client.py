from collections.abc import Mapping
import os
from typing import Iterable

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
from beneath.connection import Connection
from beneath.config import DEFAULT_READ_ALL_MAX_BYTES, DEFAULT_READ_BATCH_SIZE


class Client:
  """
  Client for interacting with Beneath.
  Data-plane features are implemented directly on Client, while control-plane features
  are isolated in the `admin` member.
  """

  def __init__(self, secret=None):
    """
    Args:
      secret (str): A beneath secret to use for authentication. If not set, reads secret from ~/.beneath.
    """
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

  async def find_stream(self, project: str = None, stream: str = None, stream_id: str = None) -> Stream:
    stream = Stream(client=self, project=project, stream=stream, stream_id=stream_id)
    # pylint: disable=protected-access
    await stream._ensure_loaded()
    return stream

  async def easy_read(
    self,
    project: str = None,
    stream: str = None,
    where: str = None,
    to_dataframe=True,
    batch_size=DEFAULT_READ_BATCH_SIZE,
    max_bytes=DEFAULT_READ_ALL_MAX_BYTES,
    warn_max=True,
  ) -> Iterable[Mapping]:
    stream = await self.find_stream(project=project, stream=stream)
    cursor = await stream.query(where=where)
    res = await cursor.fetch_all(
      max_bytes=max_bytes,
      batch_size=batch_size,
      warn_max=warn_max,
      to_dataframe=to_dataframe,
    )
    return res


class AdminClient:
  """
  AdminClient isolates control-plane features
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
