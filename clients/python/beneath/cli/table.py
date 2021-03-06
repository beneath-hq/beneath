from beneath.client import Client
from beneath.utils import ProjectIdentifier, TableIdentifier
from beneath.cli.utils import (
    async_cmd,
    pretty_print_graphql_result,
    str2bool,
    project_path_help,
    table_path_help,
)


def add_subparser(root):
    table = root.add_parser(
        "table",
        help="Create and manage tables and table instances",
        description="Create and manage tables and table instances. Tables can have multiple "
        "versions, known as instances. Learn more about tables at "
        "https://about.beneath.dev/docs/concepts/tables/",
    ).add_subparsers()

    _list = table.add_parser("list", help="List tables in a project")
    _list.set_defaults(func=async_cmd(show_list))
    _list.add_argument("project_path", type=str, help=project_path_help)

    _show = table.add_parser("show", help="Show info about a table")
    _show.set_defaults(func=async_cmd(show_table))
    _show.add_argument("table_path", type=str, help=table_path_help)

    _create = table.add_parser("create", help="Create a new table")
    _create.set_defaults(func=async_cmd(create))
    _create.add_argument("table_path", type=str, help=table_path_help)
    _create.add_argument(
        "-f",
        "--file",
        type=str,
        required=True,
        help="This file should contain the GraphQL schema for the table you would like to create",
    )
    _create.add_argument(
        "--allow-manual-writes", type=str2bool, nargs="?", const=True, default=None
    )
    _create.add_argument("--use-index", type=str2bool, nargs="?", const=True, default=None)
    _create.add_argument("--use-warehouse", type=str2bool, nargs="?", const=True, default=None)
    _create.add_argument(
        "--log-retention", type=int, help="Retention in seconds or 0 for infinite retention"
    )
    _create.add_argument(
        "--index-retention", type=int, help="Retention in seconds or 0 for infinite retention"
    )
    _create.add_argument(
        "--warehouse-retention", type=int, help="Retention in seconds or 0 for infinite retention"
    )

    _delete = table.add_parser("delete", help="Delete a table and its instances")
    _delete.set_defaults(func=async_cmd(delete))
    _delete.add_argument("table_path", type=str, help=table_path_help)

    table_instance = table.add_parser(
        "instance",
        help="Manage table instances (versions)",
        description="Manage table instances (versions). "
        "Learn more about table instances at https://about.beneath.dev/docs/concepts/tables/",
    ).add_subparsers()

    _instance_list = table_instance.add_parser("list", help="List all instances for table")
    _instance_list.set_defaults(func=async_cmd(instance_list))
    _instance_list.add_argument("table_path", type=str, help=table_path_help)

    _instance_create = table_instance.add_parser("create", help="Create new instance")
    _instance_create.set_defaults(func=async_cmd(instance_create))
    _instance_create.add_argument("table_path", type=str, help=table_path_help)
    _instance_create.add_argument("--version", type=int, default=0)
    _instance_create.add_argument(
        "--make-primary", type=str2bool, nargs="?", const=True, default=None
    )

    _instance_update = table_instance.add_parser("update", help="Update instance")
    _instance_update.set_defaults(func=async_cmd(instance_update))
    _instance_update.add_argument("instance_id", type=str)
    _instance_update.add_argument(
        "--make-final", type=str2bool, nargs="?", const=True, default=None
    )
    _instance_update.add_argument(
        "--make-primary", type=str2bool, nargs="?", const=True, default=None
    )

    _instance_clear = table_instance.add_parser("delete", help="Delete instance")
    _instance_clear.set_defaults(func=async_cmd(instance_delete))
    _instance_clear.add_argument("instance_id", type=str)


async def show_list(args):
    client = Client()
    pq = ProjectIdentifier.from_path(args.project_path)
    project = await client.admin.projects.find_by_organization_and_name(pq.organization, pq.project)
    if len(project["tables"]) == 0:
        print("There are no tables currently in this project")
    for tablename in project["tables"]:
        print(f"{pq.organization}/{pq.project}/{tablename['name']}")


async def show_table(args):
    client = Client()
    sq = TableIdentifier.from_path(args.table_path)
    table = await client.admin.tables.find_by_organization_project_and_name(
        organization_name=sq.organization,
        project_name=sq.project,
        table_name=sq.table,
    )
    _pretty_print_table(table)


async def create(args):
    with open(args.file, "r") as f:
        schema = f.read()
    client = Client()
    sq = TableIdentifier.from_path(args.table_path)
    table = await client.admin.tables.create(
        organization_name=sq.organization,
        project_name=sq.project,
        table_name=sq.table,
        schema_kind="GraphQL",
        schema=schema,
        allow_manual_writes=args.allow_manual_writes,
        use_index=args.use_index,
        use_warehouse=args.use_warehouse,
        log_retention_seconds=args.log_retention,
        index_retention_seconds=args.index_retention,
        warehouse_retention_seconds=args.warehouse_retention,
    )
    _pretty_print_table(table)


async def delete(args):
    client = Client()
    sq = TableIdentifier.from_path(args.table_path)
    table = await client.admin.tables.find_by_organization_project_and_name(
        organization_name=sq.organization,
        project_name=sq.project,
        table_name=sq.table,
    )
    result = await client.admin.tables.delete(table["tableID"])
    pretty_print_graphql_result(result)


async def instance_list(args):
    client = Client()
    sq = TableIdentifier.from_path(args.table_path)
    table = await client.admin.tables.find_by_organization_project_and_name(
        organization_name=sq.organization,
        project_name=sq.project,
        table_name=sq.table,
    )
    result = await client.admin.tables.find_instances(table["tableID"])
    pretty_print_graphql_result(result)


async def instance_create(args):
    client = Client()
    sq = TableIdentifier.from_path(args.table_path)
    table = await client.admin.tables.find_by_organization_project_and_name(
        organization_name=sq.organization,
        project_name=sq.project,
        table_name=sq.table,
    )
    result = await client.admin.tables.create_instance(
        table_id=table["tableID"],
        version=args.version,
        make_primary=args.make_primary,
    )
    pretty_print_graphql_result(result)


async def instance_update(args):
    client = Client()
    result = await client.admin.tables.update_instance(
        instance_id=args.instance_id,
        make_final=args.make_final,
        make_primary=args.make_primary,
    )
    pretty_print_graphql_result(result)


async def instance_delete(args):
    client = Client()
    result = await client.admin.tables.delete_instance(args.instance_id)
    pretty_print_graphql_result(result)


def _pretty_print_table(table):
    pretty_print_graphql_result(
        table,
        [
            "tableID",
            "name",
            "description",
            "createdOn",
            "updatedOn",
            "project",
            "schemaKind",
            "schema",
            "tableIndexes",
            "meta",
            "allowManualWrites",
            "useLog",
            "useIndex",
            "useWarehouse",
            "logRetentionSeconds",
            "indexRetentionSeconds",
            "warehouseRetentionSeconds",
            "primaryTableInstanceID",
            "instancesCreatedCount",
            "instancesDeletedCount",
            "instancesMadeFinalCount",
            "instancesMadePrimaryCount",
        ],
    )
