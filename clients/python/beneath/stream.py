# allows us to use Client as a type hint without an import cycle
# see: https://www.stefaanlippens.net/circular-imports-type-hints-python.html
# pylint: disable=wrong-import-position,ungrouped-imports
from __future__ import annotations
from typing import TYPE_CHECKING
if TYPE_CHECKING:
  from beneath.client import Client

from typing import Iterable
import uuid

from beneath.instance import DryStreamInstance, StreamInstance
from beneath.schema import Schema
from beneath.utils import StreamQualifier


class Stream:
  """
  Represents a data-plane connection to a stream. I.e., this class is used to
  query, peek, consume and write to a stream. You cannot use this class to do control-plane actions
  like creating streams or updating their details (use `client.admin.streams` directly).
  """

  client: Client
  qualifier: StreamQualifier
  admin_data: dict
  stream_id: uuid.UUID
  schema: Schema
  primary_instance: StreamInstance

  # INITIALIZATION

  def __init__(self):
    self.client: Client = None
    self.qualifier: StreamQualifier = None
    self.admin_data: dict = None
    self.stream_id: uuid.UUID = None
    self.schema: Schema = None
    self.primary_instance: StreamInstance = None

  @classmethod
  async def make(cls, client: Client, qualifier: StreamQualifier, admin_data=None) -> Stream:
    stream = Stream()
    stream.client = client
    stream.qualifier = qualifier

    if not admin_data:
      # pylint: disable=protected-access
      admin_data = await stream._load_admin_data()
    # pylint: disable=protected-access
    stream._set_admin_data(admin_data)

    return stream

  async def _load_admin_data(self):
    return await self.client.admin.streams.find_by_organization_project_and_name(
      organization_name=self.qualifier.organization,
      project_name=self.qualifier.project,
      stream_name=self.qualifier.stream,
    )

  def _set_admin_data(self, admin_data):
    self.admin_data = admin_data
    self.stream_id = uuid.UUID(hex=self.admin_data["streamID"])
    self.schema = Schema(self.admin_data["avroSchema"])
    if "primaryStreamInstance" in self.admin_data and self.admin_data["primaryStreamInstance"] is not None:
      self.primary_instance = StreamInstance(stream=self, admin_data=self.admin_data["primaryStreamInstance"])

  # MANAGEMENT

  async def delete(self):
    await self.client.admin.streams.delete(self.stream_id)

  # INSTANCES

  async def find_instances(self) -> Iterable[StreamInstance]:
    instances = await self.client.admin.streams.find_instances(self.admin_data["streamID"])
    instances = [StreamInstance(stream=self, admin_data=i) for i in instances]
    return instances

  async def stage_instance(self, version: int, make_primary=None, make_final=None, dry=False) -> StreamInstance:
    if dry:
      return DryStreamInstance(self, version=version, primary=make_primary, final=make_final)
    instance = await self.client.admin.streams.stage_instance(
      stream_id=self.admin_data["streamID"],
      version=version,
      make_final=make_final,
      make_primary=make_primary,
    )
    instance = StreamInstance(stream=self, admin_data=instance)
    return instance
