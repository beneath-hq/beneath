from beneath.client import Client
from beneath.utils import ProjectQualifier, StreamQualifier
from beneath.cli.utils import async_cmd, pretty_print_graphql_result, str2bool


def add_subparser(root):
  stream = root.add_parser('stream').add_subparsers()
  stream_instance = stream.add_parser('instance').add_subparsers()

  _list = stream.add_parser('list')
  _list.set_defaults(func=async_cmd(show_list))
  _list.add_argument('project_path', type=str)

  _show = stream.add_parser('show')
  _show.set_defaults(func=async_cmd(show_stream))
  _show.add_argument('stream_path', type=str)

  _stage = stream.add_parser('stage')
  _stage.set_defaults(func=async_cmd(stage))
  _stage.add_argument('stream_path', type=str)
  _stage.add_argument(
    '-f',
    '--file',
    type=str,
    required=True,
    help="This file should contain the GraphQL schema for the stream you would like to create",
  )
  _stage.add_argument('--allow-manual-writes', type=str2bool, nargs='?', const=True, default=None)
  _stage.add_argument('--use-index', type=str2bool, nargs='?', const=True, default=None)
  _stage.add_argument('--use-warehouse', type=str2bool, nargs='?', const=True, default=None)
  _stage.add_argument('--log-retention', type=int, help="Retention in seconds or 0 for infinite retention")
  _stage.add_argument('--index-retention', type=int, help="Retention in seconds or 0 for infinite retention")
  _stage.add_argument('--warehouse-retention', type=int, help="Retention in seconds or 0 for infinite retention")

  _delete = stream.add_parser('delete')
  _delete.set_defaults(func=async_cmd(delete))
  _delete.add_argument('stream_path', type=str)

  _instance_list = stream_instance.add_parser('list')
  _instance_list.set_defaults(func=async_cmd(instance_list))
  _instance_list.add_argument('stream_path', type=str)

  _instance_stage = stream_instance.add_parser('stage')
  _instance_stage.set_defaults(func=async_cmd(instance_stage))
  _instance_stage.add_argument('stream_path', type=str)
  _instance_stage.add_argument('--version', type=int, default=0)
  _instance_stage.add_argument('--make-final', type=str2bool, nargs='?', const=True, default=None)
  _instance_stage.add_argument('--make-primary', type=str2bool, nargs='?', const=True, default=None)

  _instance_update = stream_instance.add_parser('update')
  _instance_update.set_defaults(func=async_cmd(instance_update))
  _instance_update.add_argument('instance', type=str)
  _instance_update.add_argument('--make-final', type=str2bool, nargs='?', const=True, default=None)
  _instance_update.add_argument('--make-primary', type=str2bool, nargs='?', const=True, default=None)

  _instance_clear = stream_instance.add_parser('delete')
  _instance_clear.set_defaults(func=async_cmd(instance_delete))
  _instance_clear.add_argument('instance', type=str)


async def show_list(args):
  client = Client()
  pq = ProjectQualifier.from_path(args.project_path)
  project = await client.admin.projects.find_by_organization_and_name(pq.organization, pq.project)
  if len(project['streams']) == 0:
    print("There are no streams currently in this project")
  for streamname in project['streams']:
    print(f"{pq.organization}/{pq.project}/{streamname['name']}")


async def show_stream(args):
  client = Client()
  sq = StreamQualifier.from_path(args.stream_path)
  stream = await client.admin.streams.find_by_organization_project_and_name(
    organization_name=sq.organization,
    project_name=sq.project,
    stream_name=sq.stream,
  )
  _pretty_print_stream(stream)


async def stage(args):
  with open(args.file, "r") as f:
    schema = f.read()
  client = Client()
  sq = StreamQualifier.from_path(args.stream_path)
  stream = await client.admin.streams.stage(
    organization_name=sq.organization,
    project_name=sq.project,
    stream_name=sq.stream,
    schema_kind="GraphQL",
    schema=schema,
    allow_manual_writes=args.allow_manual_writes,
    use_index=args.use_index,
    use_warehouse=args.use_warehouse,
    log_retention_seconds=args.log_retention,
    index_retention_seconds=args.index_retention,
    warehouse_retention_seconds=args.warehouse_retention,
  )
  if not stream.get("primaryStreamInstanceID"):
    instance = await client.admin.streams.stage_instance(
      stream_id=stream["streamID"],
      version=0,
      make_primary=True,
    )
    stream["primaryStreamInstanceID"] = instance["streamInstanceID"]
    stream["primaryStreamInstance"] = instance
    stream["instancesCreatedCount"] += 1
    stream["instancesMadePrimaryCount"] += 1
  _pretty_print_stream(stream)


async def delete(args):
  client = Client()
  sq = StreamQualifier.from_path(args.stream_path)
  stream = await client.admin.streams.find_by_organization_project_and_name(
    organization_name=sq.organization,
    project_name=sq.project,
    stream_name=sq.stream,
  )
  result = await client.admin.streams.delete(stream['streamID'])
  pretty_print_graphql_result(result)


async def instance_list(args):
  client = Client()
  sq = StreamQualifier.from_path(args.stream_path)
  stream = await client.admin.streams.find_by_organization_project_and_name(
    organization_name=sq.organization,
    project_name=sq.project,
    stream_name=sq.stream,
  )
  result = await client.admin.streams.find_instances(stream['streamID'])
  pretty_print_graphql_result(result)


async def instance_stage(args):
  client = Client()
  sq = StreamQualifier.from_path(args.stream_path)
  stream = await client.admin.streams.find_by_organization_project_and_name(
    organization_name=sq.organization,
    project_name=sq.project,
    stream_name=sq.stream,
  )
  result = await client.admin.streams.stage_instance(
    stream_id=stream['streamID'],
    version=args.version,
    make_final=args.make_final,
    make_primary=args.make_primary,
  )
  pretty_print_graphql_result(result)


async def instance_update(args):
  client = Client()
  result = await client.admin.streams.update_instance(
    instance_id=args.instance,
    make_final=args.make_final,
    make_primary=args.make_primary,
  )
  pretty_print_graphql_result(result)


async def instance_delete(args):
  client = Client()
  result = await client.admin.streams.delete_instance(args.instance)
  pretty_print_graphql_result(result)


def _pretty_print_stream(stream):
  pretty_print_graphql_result(
    stream, [
      "streamID",
      "name",
      "description",
      "createdOn",
      "updatedOn",
      "project",
      "schemaKind",
      "schema",
      "streamIndexes",
      "allowManualWrites",
      "useLog",
      "useIndex",
      "useWarehouse",
      "logRetentionSeconds",
      "indexRetentionSeconds",
      "warehouseRetentionSeconds",
      "primaryStreamInstanceID",
      "instancesCreatedCount",
      "instancesDeletedCount",
      "instancesMadeFinalCount",
      "instancesMadePrimaryCount",
    ]
  )
