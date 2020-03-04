import os

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
    await stream.ensure_loaded()
    return stream


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
