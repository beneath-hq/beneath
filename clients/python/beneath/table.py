# allows us to use Client as a type hint without an import cycle
# see: https://www.stefaanlippens.net/circular-imports-type-hints-python.html
# pylint: disable=wrong-import-position,ungrouped-imports
from __future__ import annotations
from typing import TYPE_CHECKING

if TYPE_CHECKING:
    from beneath.client import Client

from typing import Iterable, Mapping, Union
import uuid

from beneath import config
from beneath.cursor import Cursor
from beneath.instance import TableInstance
from beneath.schema import Schema
from beneath.utils import TableQualifier


class Table:
    """
    Represents a data-plane connection to a table.
    To find or create a table, see :class:`beneath.Client`.

    Use it to get a TableInstance, which you can query, replay, subscribe and write to.
    Learn more about tables and instances at https://about.beneath.dev/docs/concepts/tables/.
    """

    table_id: uuid.UUID
    """
    The table ID
    """

    schema: Schema
    """
    The table's schema
    """

    primary_instance: TableInstance
    """
    The current primary table instance.
    This is probably the object you will use to write/query the table.
    """

    use_log: bool
    """ Whether log queries are supported for this table. """

    use_index: bool
    """ Whether index queries are supported for this table. """

    use_warehouse: bool
    """ Whether warehouse queries are supported for this table. """

    _client: Client
    _qualifier: TableQualifier

    # INITIALIZATION

    def __init__(self):
        self.table_id: uuid.UUID = None
        self.schema: Schema = None
        self.primary_instance: TableInstance = None
        self._client: Client = None
        self._qualifier: TableQualifier = None

    @classmethod
    async def _make(cls, client: Client, qualifier: TableQualifier, admin_data=None) -> Table:
        table = Table()
        table._client = client
        table._qualifier = qualifier
        if not admin_data:
            # pylint: disable=protected-access
            admin_data = await table._load_admin_data()
        table.table_id = uuid.UUID(hex=admin_data["tableID"])
        table.schema = Schema(admin_data["avroSchema"])
        if "primaryTableInstance" in admin_data:
            if admin_data["primaryTableInstance"] is not None:
                table.primary_instance = TableInstance._make(
                    client=client,
                    table=table,
                    admin_data=admin_data["primaryTableInstance"],
                )
        table.use_log = admin_data["useLog"]
        table.use_index = admin_data["useIndex"]
        table.use_warehouse = admin_data["useWarehouse"]
        return table

    @classmethod
    async def _make_dry(
        cls,
        client: Client,
        qualifier: TableQualifier,
        avro_schema: str,
    ) -> Table:
        table = Table()
        table._client = client
        table._qualifier = qualifier
        table.table_id = None
        table.schema = Schema(avro_schema)
        table.primary_instance = await table.create_instance(version=0, make_primary=True)
        table.use_log = True
        table.use_index = True
        table.use_warehouse = True
        return table

    async def _load_admin_data(self):
        return await self._client.admin.tables.find_by_organization_project_and_name(
            organization_name=self._qualifier.organization,
            project_name=self._qualifier.project,
            table_name=self._qualifier.table,
        )

    # STATE

    def __repr__(self):
        return f'<beneath.table.Table("{config.BENEATH_FRONTEND_HOST}/{self._qualifier}")>'

    # INSTANCES

    async def find_instances(self) -> Iterable[TableInstance]:
        """
        Returns a list of all the table's instances.
        Learn more about instances at https://about.beneath.dev/docs/concepts/tables/.
        """
        # handle if dry
        if not self.table_id:
            if self.primary_instance:
                return [self.primary_instance]
            else:
                return []
        instances = await self._client.admin.tables.find_instances(str(self.table_id))
        instances = [
            TableInstance._make(client=self._client, table=self, admin_data=i) for i in instances
        ]
        return instances

    async def find_instance(self, version: int):
        """
        Finds an instance by version number
        Learn more about instances at https://about.beneath.dev/docs/concepts/tables/.
        """
        # handle dry case
        if not self.table_id:
            if self.primary_instance and self.primary_instance.version == version:
                return self.primary_instance
            raise Exception("can't find instance by version for table created with a dry client")
        if self.primary_instance and self.primary_instance.version == version:
            return self.primary_instance
        admin_data = await self._client.admin.tables.find_instance(
            table_id=str(self.table_id),
            version=version,
        )
        instance = TableInstance._make(client=self._client, table=self, admin_data=admin_data)
        return instance

    async def create_instance(
        self,
        version: int = None,
        make_primary=None,
        update_if_exists=None,
    ) -> TableInstance:
        """
        Creates and returns a new instance for the table.
        Learn more about instances at https://about.beneath.dev/docs/concepts/tables/.

        Args:
            version (int):
                The version number to assign to the instance. If not set, will create a new instance
                with a higher version number than any previous instance for the table.
            make_primary (bool):
                Immediately make the new instance the table's primary instance
            update_if_exists (bool):
                If true and an instance for ``version`` already exists, will update and return the
                existing instance.
        """
        # handle real and dry cases
        if self.table_id:
            admin_data = await self._client.admin.tables.create_instance(
                table_id=str(self.table_id),
                version=version,
                make_primary=make_primary,
                update_if_exists=update_if_exists,
            )
            instance = TableInstance._make(client=self._client, table=self, admin_data=admin_data)
        else:
            instance = TableInstance._make_dry(
                client=self._client,
                table=self,
                version=(0 if version is None else version),
                make_primary=make_primary,
            )
        if make_primary:
            self.primary_instance = instance
        return instance

    # MANAGEMENT

    async def delete(self):
        """
        Deletes the table and all its instances and data.
        """
        # handle if dry
        if not self.table_id:
            raise Exception("cannot delete dry table")
        await self._client.admin.tables.delete(self.table_id)

    # CURSORS

    def restore_cursor(self, replay_cursor: bytes, changes_cursor: bytes):
        """
        Restores a cursor previously obtained by querying one of the table's instances.
        You must provide the cursor bytes, which can be found as properties of the Cursor object.
        """
        return Cursor(
            connection=self._client.connection,
            schema=self.schema,
            replay_cursor=replay_cursor,
            changes_cursor=changes_cursor,
        )

    # WRITING RECORDS

    async def write(self, records: Union[Mapping, Iterable[Mapping]]):
        """
        Writes records to the table's primary instance.

        This is a convenience wrapper for ``table.primary_instance.write(...)``.
        """
        if not self.primary_instance:
            raise Exception("cannot write because the table doesn't have a primary instance")
        await self.primary_instance.write(records)
