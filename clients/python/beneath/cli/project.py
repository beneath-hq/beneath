from beneath.client import Client
from beneath.utils import ProjectQualifier
from beneath.cli.utils import async_cmd, pretty_print_graphql_result, str2bool


def add_subparser(root):
  project = root.add_parser('project').add_subparsers()

  _stage = project.add_parser('stage')
  _stage.set_defaults(func=async_cmd(stage))
  _stage.add_argument('project_path', type=str)
  _stage.add_argument('--display-name', type=str)
  _stage.add_argument('--public', type=str2bool, nargs='?', const=True, default=None)
  _stage.add_argument('--description', type=str)
  _stage.add_argument('--site-url', type=str)
  _stage.add_argument('--photo-url', type=str)

  _delete = project.add_parser('delete')
  _delete.set_defaults(func=async_cmd(delete))
  _delete.add_argument('project_path', type=str)

  _transfer = project.add_parser('transfer')
  _transfer.set_defaults(func=async_cmd(transfer_organization))
  _transfer.add_argument('project_path', type=str)
  _transfer.add_argument('new_organization', type=str)

  _member_permissions = project.add_parser('show-members')
  _member_permissions.set_defaults(func=async_cmd(show_member_permissions))
  _member_permissions.add_argument('project_path', type=str)

  _update_member_permissions = project.add_parser('update-permissions')
  _update_member_permissions.set_defaults(func=async_cmd(update_member_permissions))
  _update_member_permissions.add_argument('project_path', type=str)
  _update_member_permissions.add_argument('username', type=str)
  _update_member_permissions.add_argument(
    '--view',
    type=str2bool,
    nargs='?',
    const=True,
    default=None,
  )
  _update_member_permissions.add_argument(
    '--create',
    type=str2bool,
    nargs='?',
    const=True,
    default=None,
  )
  _update_member_permissions.add_argument(
    '--admin',
    type=str2bool,
    nargs='?',
    const=True,
    default=None,
  )

async def stage(args):
  client = Client()
  pq = ProjectQualifier.from_path(args.project_path)
  result = await client.admin.projects.stage(
    organization_name=pq.organization,
    project_name=pq.project,
    display_name=args.display_name,
    public=args.public,
    description=args.description,
    site_url=args.site_url,
    photo_url=args.photo_url,
  )
  pretty_print_graphql_result(result)


async def delete(args):
  client = Client()
  pq = ProjectQualifier.from_path(args.project_path)
  project = await client.admin.projects.find_by_organization_and_name(pq.organization, pq.project)
  result = await client.admin.projects.delete(project_id=project['projectID'])
  pretty_print_graphql_result(result)


async def transfer_organization(args):
  client = Client()
  pq = ProjectQualifier.from_path(args.project_path)
  project = await client.admin.projects.find_by_organization_and_name(pq.organization, pq.project)
  organization = await client.admin.organizations.find_by_name(args.new_organization)
  result = await client.admin.organizations.transfer_project(
    project_id=project["projectID"],
    new_organization_id=organization["organizationID"],
  )
  pretty_print_graphql_result(result)


async def show_member_permissions(args):
  client = Client()
  pq = ProjectQualifier.from_path(args.project_path)
  project = await client.admin.projects.find_by_organization_and_name(pq.organization, pq.project)
  result = await client.admin.projects.get_member_permissions(project_id=project['projectID'])
  pretty_print_graphql_result(result)


async def update_member_permissions(args):
  client = Client()
  pq = ProjectQualifier.from_path(args.project_path)
  project = await client.admin.projects.find_by_organization_and_name(pq.organization, pq.project)
  user = await client.admin.organizations.find_by_name(args.username)
  result = await client.admin.users.update_permissions_for_project(
    user_id=user["personalUserID"],
    project_id=project["projectID"],
    view=args.view,
    create=args.create,
    admin=args.admin,
  )
  pretty_print_graphql_result(result)
