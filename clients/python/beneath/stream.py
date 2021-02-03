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
    Represents a data-plane connection to a stream.
    To find or create a stream, see :class:`beneath.Client`.

    Use it to get a StreamInstance, which you can query, replay, subscribe and write to.
    Learn more about streams and instances at https://about.beneath.dev/docs/concepts/streams/.
    """

    client: Client

    qualifier: StreamQualifier

    admin_data: dict

    stream_id: uuid.UUID
    """
    The stream ID
    """

    schema: Schema
    """
    The stream's schema
    """

    primary_instance: StreamInstance
    """
    The current primary stream instance.
    This is probably the object you will use to write/query the stream.
    """

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
        if (
            "primaryStreamInstance" in self.admin_data
            and self.admin_data["primaryStreamInstance"] is not None
        ):
            self.primary_instance = StreamInstance(
                stream=self, admin_data=self.admin_data["primaryStreamInstance"]
            )

    # MANAGEMENT

    async def delete(self):
        """
        Deletes the stream and all its instances and data.
        """
        await self.client.admin.streams.delete(self.stream_id)

    # INSTANCES

    async def find_instances(self) -> Iterable[StreamInstance]:
        """
        Returns a list of all the stream's instances.
        Learn more about instances at https://about.beneath.dev/docs/concepts/streams/.
        """
        instances = await self.client.admin.streams.find_instances(self.admin_data["streamID"])
        instances = [StreamInstance(stream=self, admin_data=i) for i in instances]
        return instances

    async def create_instance(
        self,
        version: int,
        make_primary=None,
        update_if_exists=None,
        dry=False,
    ) -> StreamInstance:
        """
        Creates and returns a new instance for the stream.
        Learn more about instances at https://about.beneath.dev/docs/concepts/streams/.

        Args:
            version (int):
                The version number to assign to the instance
            make_primary (bool):
                Immediately make the new instance the stream's primary instance
            update_if_exists (bool):
                If true and an instance for ``version`` already exists, will update and return the
                existing instance.
        """
        if dry:
            return DryStreamInstance(self, version=version, primary=make_primary, final=None)
        instance = await self.client.admin.streams.create_instance(
            stream_id=self.admin_data["streamID"],
            version=version,
            make_primary=make_primary,
            update_if_exists=update_if_exists,
        )
        instance = StreamInstance(stream=self, admin_data=instance)
        return instance
