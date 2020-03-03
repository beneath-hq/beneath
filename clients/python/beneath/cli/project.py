from beneath.client import Client
from beneath.cli.utils import async_cmd, parse_names, pretty_print_graphql_result, str2bool


def add_subparser(root):
  project = root.add_parser('project').add_subparsers()

  _list = project.add_parser('list')
  _list.set_defaults(func=async_cmd(show_list))

  _create = project.add_parser('create')
  _create.set_defaults(func=async_cmd(create))
  _create.add_argument('name', type=str)
  _create.add_argument('--display-name', type=str)
  _create.add_argument('-o', '--organization', type=str)
  _create.add_argument('--public', type=str2bool, nargs='?', const=True, default=True)
  _create.add_argument('--description', type=str)
  _create.add_argument('--site-url', type=str)
  _create.add_argument('--photo-url', type=str)

  _update = project.add_parser('update')
  _update.set_defaults(func=async_cmd(update))
  _update.add_argument('name', type=str)
  _update.add_argument('--display-name', type=str)
  _update.add_argument('--public', type=str2bool, nargs='?', const=True, default=None)
  _update.add_argument('--description', type=str)
  _update.add_argument('--site-url', type=str)
  _update.add_argument('--photo-url', type=str)

  _delete = project.add_parser('delete')
  _delete.set_defaults(func=async_cmd(delete))
  _delete.add_argument('name', type=str)

  _add_user = project.add_parser('add-member')
  _add_user.set_defaults(func=async_cmd(add_member))
  _add_user.add_argument('project', type=str)
  _add_user.add_argument('username', type=str)
  _add_user.add_argument('--view', type=str2bool, nargs='?', const=True, default=True)
  _add_user.add_argument('--create', type=str2bool, nargs='?', const=True, default=True)
  _add_user.add_argument('--admin', type=str2bool, nargs='?', const=True, default=False)

  _remove_user = project.add_parser('remove-member')
  _remove_user.set_defaults(func=async_cmd(remove_member))
  _remove_user.add_argument('project', type=str)
  _remove_user.add_argument('username', type=str)

  _migrate = project.add_parser('migrate-organization')
  _migrate.set_defaults(func=async_cmd(migrate_organization))
  _migrate.add_argument('project', type=str)
  _migrate.add_argument('--new-organization', type=str, required=True)


async def show_list(args):
  client = Client()
  me = await client.admin.users.get_me()
  user = await client.admin.users.get_by_id(me['userID'])
  for project in user['projects']:
    print(project['name'])


async def create(args):
  client = Client()
  name, org_name = parse_names(args.name, args.organization, "organization")
  organization = await client.admin.organizations.find_by_name(org_name)
  result = await client.admin.projects.create(
    name=name,
    display_name=args.display_name,
    organization_id=organization['organizationID'],
    public=args.public,
    description=args.description,
    site_url=args.site_url,
    photo_url=args.photo_url,
  )
  pretty_print_graphql_result(result)


async def update(args):
  client = Client()
  project = await client.admin.projects.find_by_name(args.name)
  result = await client.admin.projects.update_details(
    project_id=project['projectID'],
    display_name=args.display_name,
    public=args.public,
    description=args.description,
    site_url=args.site_url,
    photo_url=args.photo_url,
  )
  pretty_print_graphql_result(result)


async def delete(args):
  client = Client()
  project = await client.admin.projects.find_by_name(args.name)
  result = await client.admin.projects.delete(project_id=project['projectID'])
  pretty_print_graphql_result(result)


async def add_member(args):
  client = Client()
  project = await client.admin.projects.find_by_name(args.project)
  result = await client.admin.projects.add_user(
    project_id=project['projectID'],
    username=args.username,
    view=args.view,
    create=args.create,
    admin=args.admin,
  )
  pretty_print_graphql_result(result)


async def remove_member(args):
  client = Client()
  project = await client.admin.projects.find_by_name(args.project)
  user = await client.admin.users.get_by_username(args.username)
  result = await client.admin.projects.remove_user(
    project_id=project['projectID'],
    user_id=user['userID'],
  )
  pretty_print_graphql_result(result)


async def migrate_organization(args):
  client = Client()
  project = await client.admin.projects.find_by_name(args.project)
  organization = await client.admin.organizations.find_by_name(args.new_organization)
  result = await client.admin.projects.update_organization(
    project_id=project['projectID'],
    organization_id=organization['organizationID'],
  )
  pretty_print_graphql_result(result)
