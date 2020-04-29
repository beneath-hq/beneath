from beneath.client import Client
from beneath.utils import ProjectQualifier, StreamQualifier
from beneath.cli.utils import async_cmd, pretty_print_graphql_result, str2bool


def add_subparser(root):
  stream = root.add_parser('stream').add_subparsers()
  stream_batch = stream.add_parser('batch').add_subparsers()

  _list = stream.add_parser('list')
  _list.set_defaults(func=async_cmd(show_list))
  _list.add_argument('project_path', type=str)

  _create = stream.add_parser('create')
  _create.set_defaults(func=async_cmd(create_root))
  _create.add_argument(
    '-f',
    '--file',
    type=str,
    required=True,
    help="This file should contain the GraphQL schema for the stream you would like to create"
  )
  _create.add_argument('-p', '--project-path', type=str, required=True)
  _create.add_argument('--manual', type=str2bool, nargs='?', const=True, default=False)
  _create.add_argument('--batch', type=str2bool, nargs='?', const=True, default=False)

  _update = stream.add_parser('update')
  _update.set_defaults(func=async_cmd(update_root))
  _update.add_argument('stream_path', type=str)
  _update.add_argument(
    '-f',
    '--file',
    type=str,
    required=True,
    help=
    "This file should contain the stream's schema and an updated description. Only the description should change, the schema itself should not change."
  )
  _update.add_argument('--manual', type=str2bool, nargs='?', const=True, default=False)

  _delete = stream.add_parser('delete')
  _delete.set_defaults(func=async_cmd(delete_root))
  _delete.add_argument('stream_path', type=str)

  _batch_create = stream_batch.add_parser('create')
  _batch_create.set_defaults(func=async_cmd(batch_create))
  _batch_create.add_argument('stream_path', type=str)

  _batch_commit = stream_batch.add_parser('commit')
  _batch_commit.set_defaults(func=async_cmd(batch_commit))
  _batch_commit.add_argument('instance', type=str)

  _batches_clear = stream_batch.add_parser('clear')
  _batches_clear.set_defaults(func=async_cmd(batches_clear))
  _batches_clear.add_argument('stream_path', type=str)


async def show_list(args):
  client = Client()
  pq = ProjectQualifier.from_path(args.project_path)
  project = await client.admin.projects.find_by_organization_and_name(pq.organization, pq.project)
  if len(project['streams']) == 0:
    print("There are no streams currently in this project")
  for streamname in project['streams']:
    print(f"{pq.organization}/{pq.project}/{streamname['name']}")


async def create_root(args):
  with open(args.file, "r") as f:
    schema = f.read()

  client = Client()
  pq = ProjectQualifier.from_path(args.project_path)
  project = await client.admin.projects.find_by_organization_and_name(pq.organization, pq.project)
  stream = await client.admin.streams.create(
    schema=schema,
    project_id=project['projectID'],
    manual=args.manual,
    batch=args.batch,
  )

  _pretty_print_stream(stream)


async def update_root(args):
  schema = None
  if args.file:
    with open(args.file, "r") as f:
      schema = f.read()

  client = Client()
  sq = StreamQualifier.from_path(args.stream_path)
  stream = await client.admin.streams.find_by_organization_project_and_name(
    organization_name=sq.organization,
    project_name=sq.project,
    stream_name=sq.stream,
  )

  stream = await client.admin.streams.update(stream['streamID'], schema=schema, manual=args.manual)
  _pretty_print_stream(stream)


async def delete_root(args):
  client = Client()
  sq = StreamQualifier.from_path(args.stream_path)
  stream = await client.admin.streams.find_by_organization_project_and_name(
    organization_name=sq.organization,
    project_name=sq.project,
    stream_name=sq.stream,
  )
  result = await client.admin.streams.delete(stream['streamID'])
  pretty_print_graphql_result(result)


async def batch_create(args):
  client = Client()
  sq = StreamQualifier.from_path(args.stream_path)
  stream = await client.admin.streams.find_by_organization_project_and_name(
    organization_name=sq.organization,
    project_name=sq.project,
    stream_name=sq.stream,
  )
  batch = await client.admin.streams.create_batch(stream['streamID'])
  pretty_print_graphql_result(batch)


async def batch_commit(args):
  client = Client()
  result = await client.admin.streams.commit_batch(instance_id=args.instance)
  pretty_print_graphql_result(result)


async def batches_clear(args):
  client = Client()
  sq = StreamQualifier.from_path(args.stream_path)
  stream = await client.admin.streams.find_by_organization_project_and_name(
    organization_name=sq.organization,
    project_name=sq.project,
    stream_name=sq.stream,
  )
  result = await client.admin.streams.clear_pending_batches(stream['streamID'])
  pretty_print_graphql_result(result)


def _pretty_print_stream(stream):
  pretty_print_graphql_result(
    stream, [
      "streamID",
      "name",
      "description",
      "createdOn",
      "updatedOn",
      "external",
      "batch",
      "manual",
      "retentionSeconds"
      "project",
      "streamIndexes",
      "currentStreamInstanceID",
    ]
  )
