from beneath.client import Client
from beneath.cli.utils import async_cmd, parse_names, pretty_print_graphql_result, str2bool


def add_subparser(root):
  stream = root.add_parser('stream').add_subparsers()
  root_stream = root.add_parser('root-stream').add_subparsers()
  root_stream_batch = root_stream.add_parser('batch').add_subparsers()

  _list = stream.add_parser('list')
  _list.set_defaults(func=async_cmd(show_list))
  _list.add_argument('project', type=str)

  _create = root_stream.add_parser('create')
  _create.set_defaults(func=async_cmd(create_root))
  _create.add_argument(
    '-f',
    '--file',
    type=str,
    required=True,
    help="This file should contain the GraphQL schema for the stream you would like to create"
  )
  _create.add_argument('-p', '--project', type=str, required=True)
  _create.add_argument('--manual', type=str2bool, nargs='?', const=True, default=False)
  _create.add_argument('--batch', type=str2bool, nargs='?', const=True, default=False)

  _update = root_stream.add_parser('update')
  _update.set_defaults(func=async_cmd(update_root))
  _update.add_argument('stream', type=str)
  _update.add_argument('-p', '--project', type=str)
  _update.add_argument(
    '-f',
    '--file',
    type=str,
    required=True,
    help=
    "This file should contain the stream's schema and an updated description. Only the description should change, the schema itself should not change."
  )
  _update.add_argument('--manual', type=str2bool, nargs='?', const=True, default=False)

  _delete = root_stream.add_parser('delete')
  _delete.set_defaults(func=async_cmd(delete_root))
  _delete.add_argument('stream', type=str)
  _delete.add_argument('-p', '--project', type=str)

  _batch_create = root_stream_batch.add_parser('create')
  _batch_create.set_defaults(func=async_cmd(batch_create))
  _batch_create.add_argument('stream', type=str)
  _batch_create.add_argument('-p', '--project', type=str)

  _batch_commit = root_stream_batch.add_parser('commit')
  _batch_commit.set_defaults(func=async_cmd(batch_commit))
  _batch_commit.add_argument('instance', type=str)

  _batches_clear = root_stream_batch.add_parser('clear')
  _batches_clear.set_defaults(func=async_cmd(batches_clear))
  _batches_clear.add_argument('stream', type=str)
  _batches_clear.add_argument('-p', '--project', type=str)


async def show_list(args):
  client = Client()
  project = await client.admin.projects.find_by_name(args.project)
  if len(project['streams']) == 0:
    print("There are no streams currently in this project")
  for streamname in project['streams']:
    print(streamname['name'])


async def create_root(args):
  with open(args.file, "r") as f:
    schema = f.read()

  client = Client()
  project = await client.admin.projects.find_by_name(args.project)
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
  name, project_name = parse_names(args.stream, args.project, "project")
  stream = await client.admin.streams.find_by_project_and_name(
    project_name=project_name, stream_name=name
  )

  stream = await client.admin.streams.update(stream['streamID'], schema=schema, manual=args.manual)
  _pretty_print_stream(stream)


async def delete_root(args):
  client = Client()
  name, project_name = parse_names(args.stream, args.project, "project")
  stream = await client.admin.streams.find_by_project_and_name(
    project_name=project_name,
    stream_name=name,
  )
  result = await client.admin.streams.delete(stream['streamID'])
  pretty_print_graphql_result(result)


async def batch_create(args):
  client = Client()
  name, project_name = parse_names(args.stream, args.project, "project")
  stream = await client.admin.streams.find_by_project_and_name(
    project_name=project_name, stream_name=name
  )
  batch = await client.admin.streams.create_batch(stream['streamID'])
  pretty_print_graphql_result(batch)


async def batch_commit(args):
  client = Client()
  result = await client.admin.streams.commit_batch(instance_id=args.instance)
  pretty_print_graphql_result(result)


async def batches_clear(args):
  client = Client()
  name, project_name = parse_names(args.stream, args.project, "project")
  stream = await client.admin.streams.find_by_project_and_name(
    project_name=project_name, stream_name=name
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
