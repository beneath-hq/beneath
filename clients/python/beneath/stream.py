# allows us to use Client as a type hint without an import cycle
# see: https://www.stefaanlippens.net/circular-imports-type-hints-python.html
# pylint: disable=wrong-import-position,ungrouped-imports
from __future__ import annotations
from typing import TYPE_CHECKING

if TYPE_CHECKING:
    from beneath.client import Client

from typing import Iterable
import uuid

from beneath.instance import StreamInstance
from beneath.schema import Schema
from beneath.utils import StreamQualifier


class Stream:
    """
    Represents a data-plane connection to a stream.
    To find or create a stream, see :class:`beneath.Client`.

    Use it to get a StreamInstance, which you can query, replay, subscribe and write to.
    Learn more about streams and instances at https://about.beneath.dev/docs/concepts/streams/.
    """

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

    _client: Client
    _qualifier: StreamQualifier

    # INITIALIZATION

    def __init__(self):
        self.stream_id: uuid.UUID = None
        self.schema: Schema = None
        self.primary_instance: StreamInstance = None
        self._client: Client = None
        self._qualifier: StreamQualifier = None

    @classmethod
    async def _make(cls, client: Client, qualifier: StreamQualifier, admin_data=None) -> Stream:
        stream = Stream()
        stream._client = client
        stream._qualifier = qualifier
        if not admin_data:
            # pylint: disable=protected-access
            admin_data = await stream._load_admin_data()
        stream.stream_id = uuid.UUID(hex=admin_data["streamID"])
        stream.schema = Schema(admin_data["avroSchema"])
        if "primaryStreamInstance" in admin_data:
            if admin_data["primaryStreamInstance"] is not None:
                stream.primary_instance = StreamInstance._make(
                    client=client,
                    stream=stream,
                    admin_data=admin_data["primaryStreamInstance"],
                )
        return stream

    @classmethod
    async def _make_dry(
        cls,
        client: Client,
        qualifier: StreamQualifier,
        avro_schema: str,
    ) -> Stream:
        stream = Stream()
        stream._client = client
        stream._qualifier = qualifier
        stream.stream_id = None
        stream.schema = Schema(avro_schema)
        stream.primary_instance = await stream.create_instance(version=0, make_primary=True)
        return stream

    async def _load_admin_data(self):
        return await self._client.admin.streams.find_by_organization_project_and_name(
            organization_name=self._qualifier.organization,
            project_name=self._qualifier.project,
            stream_name=self._qualifier.stream,
        )

    # INSTANCES

    async def find_instances(self) -> Iterable[StreamInstance]:
        """
        Returns a list of all the stream's instances.
        Learn more about instances at https://about.beneath.dev/docs/concepts/streams/.
        """
        # handle if dry
        if not self.stream_id:
            if self.primary_instance:
                return [self.primary_instance]
            else:
                return []
        instances = await self._client.admin.streams.find_instances(str(self.stream_id))
        instances = [
            StreamInstance._make(client=self._client, stream=self, admin_data=i) for i in instances
        ]
        return instances

    async def find_instance(self, version: int):
        """
        Finds an instance by version number
        Learn more about instances at https://about.beneath.dev/docs/concepts/streams/.
        """
        # handle dry case
        if not self.stream_id:
            if self.primary_instance and self.primary_instance.version == version:
                return self.primary_instance
            raise Exception("can't find instance by version for stream created with a dry client")
        admin_data = await self._client.admin.streams.find_instance(
            stream_id=str(self.stream_id),
            version=version,
        )
        instance = StreamInstance._make(client=self._client, stream=self, admin_data=admin_data)
        return instance

    async def create_instance(
        self,
        version: int,
        make_primary=None,
        update_if_exists=None,
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
        # handle real and dry cases
        if self.stream_id:
            admin_data = await self._client.admin.streams.create_instance(
                stream_id=str(self.stream_id),
                version=version,
                make_primary=make_primary,
                update_if_exists=update_if_exists,
            )
            instance = StreamInstance._make(client=self._client, stream=self, admin_data=admin_data)
        else:
            instance = StreamInstance._make_dry(
                client=self._client,
                stream=self,
                version=version,
                make_primary=make_primary,
            )
        if make_primary:
            self.primary_instance = instance
        return instance

    # MANAGEMENT

    async def delete(self):
        """
        Deletes the stream and all its instances and data.
        """
        # handle if dry
        if not self.stream_id:
            raise Exception("cannot delete dry stream")
        await self._client.admin.streams.delete(self.stream_id)
