# Allows us to use Table as a type hint without an import cycle
# pylint: disable=wrong-import-position,ungrouped-imports
from __future__ import annotations
from typing import TYPE_CHECKING

if TYPE_CHECKING:
    from beneath.client import Client
    from beneath.table import Table

from collections.abc import Mapping
from typing import Iterable, Union
import uuid

from beneath import config
from beneath.cursor import Cursor


class TableInstance:
    """
    Represents an instance of a Table, i.e. a specific version that you can query/subscribe/write
    to. Learn more about instances at https://about.beneath.dev/docs/concepts/tables/.
    """

    # INITIALIZATION

    def __init__(self):
        self.table: Table = None
        """ The table that this is an instance of """

        self.instance_id: uuid.UUID = None
        """ The table instance ID """

        self.is_final: bool = None
        """ True if the instance has been made final and is closed for further writes """

        self.is_primary: bool = None
        """ True if the instance is the primary instance for the table """

        self.version: int = None
        """ The instance's version number """

        self._client: Client = None

    @classmethod
    def _make(cls, client: Client, table: Table, admin_data: dict) -> TableInstance:
        instance = TableInstance()
        instance._client = client
        instance.table = table
        instance._set_admin_data(admin_data)
        return instance

    @classmethod
    def _make_dry(
        cls,
        client: Client,
        table: Table,
        version: int,
        make_primary=False,
    ) -> TableInstance:
        instance = TableInstance()
        instance._client = client
        instance.table = table
        instance.instance_id = None
        instance.version = version
        instance.is_primary = make_primary
        return instance

    def _set_admin_data(self, admin_data: dict):
        self.instance_id = uuid.UUID(hex=admin_data["tableInstanceID"])
        self.version = admin_data["version"]
        self.is_final = admin_data["madeFinalOn"] is not None
        self.is_primary = admin_data["madePrimaryOn"] is not None

    # STATE

    def __repr__(self):
        url = f"{config.BENEATH_FRONTEND_HOST}/{self.table._qualifier}/{self.version}"
        return f'<beneath.table.TableInstance("{url}")>'

    # CONTROL PLANE

    async def update(self, make_primary=None, make_final=None):
        """ Updates the instance """
        # handle real and dry cases
        if self.instance_id:
            if make_final:
                await self._client.force_flush()
            admin_data = await self._client.admin.tables.update_instance(
                instance_id=str(self.instance_id),
                make_primary=make_primary,
                make_final=make_final,
            )
            self._set_admin_data(admin_data)
        else:
            self.is_final = self.is_final or make_final
            self.is_primary = self.is_primary or make_primary
        if make_primary:
            self.table.primary_instance = self

    async def delete(self):
        """ Deletes the instance """
        if self.instance_id:
            await self._client.admin.tables.delete_instance(instance_id=str(self.instance_id))
        if self.table.primary_instance == self:
            self.table.primary_instance = None

    # READING RECORDS

    async def query_log(self, peek: bool = False) -> Cursor:
        """
        Queries the table's log, returning a cursor for replaying every record
        written to the instance or for subscribing to new changes records in the table.

        Args:
            peek (bool):
                If true, returns a cursor for the most recent records and
                lets you page through the log in reverse order.
        """
        # handle dry case
        if not self.instance_id:
            raise Exception("cannot query a dry instance")
        resp = await self._client.connection.query_log(instance_id=self.instance_id, peek=peek)
        assert len(resp.replay_cursors) <= 1 and len(resp.change_cursors) <= 1
        replay = resp.replay_cursors[0] if len(resp.replay_cursors) > 0 else None
        changes = resp.change_cursors[0] if len(resp.change_cursors) > 0 else None
        return Cursor(
            connection=self._client.connection,
            schema=self.table.schema,
            replay_cursor=replay,
            changes_cursor=changes,
        )

    # pylint: disable=redefined-builtin
    async def query_index(self, filter: str = None) -> Cursor:
        """
        Queries a sorted index of the records written to the table. The index contains
        the newest record for each record key (see the table's schema for the key). Returns
        a cursor for paging through the index.

        Args:
            filter (str):
                A filter to apply to the index. Filters allow you to quickly
                find specific record(s) in the index based on the record key.
                For details on the filter syntax,
                see https://about.beneath.dev/docs/reading-writing-data/index-filters/.
        """
        # handle dry case
        if not self.instance_id:
            raise Exception("cannot query a dry instance")
        resp = await self._client.connection.query_index(
            instance_id=self.instance_id, filter=filter
        )
        assert len(resp.replay_cursors) <= 1 and len(resp.change_cursors) <= 1
        replay = resp.replay_cursors[0] if len(resp.replay_cursors) > 0 else None
        changes = resp.change_cursors[0] if len(resp.change_cursors) > 0 else None
        return Cursor(
            connection=self._client.connection,
            schema=self.table.schema,
            replay_cursor=replay,
            changes_cursor=changes,
        )

    # WRITING RECORDS

    async def write(self, records: Union[Mapping, Iterable[Mapping]]):
        """ Convenience wrapper for ``Client.write`` """
        await self._client.write(self, records)
