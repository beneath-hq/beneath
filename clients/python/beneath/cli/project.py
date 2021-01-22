from beneath.client import Client
from beneath.utils import pretty_entity_name, ProjectQualifier
from beneath.cli.utils import async_cmd, pretty_print_graphql_result, str2bool, project_path_help


def add_subparser(root):
    project = root.add_parser(
        "project",
        help="Browse, create and manage projects",
        description="Browse, create and manage projects",
    ).add_subparsers()

    _list = project.add_parser("list", help="List the projects you own or have access to")
    _list.set_defaults(func=async_cmd(list_for_me))

    _show = project.add_parser("show", help="Show info about a project")
    _show.set_defaults(func=async_cmd(show))
    _show.add_argument("project_path", type=str, help=project_path_help)

    _member_permissions = project.add_parser("show-members", help="Show project members")
    _member_permissions.set_defaults(func=async_cmd(show_member_permissions))
    _member_permissions.add_argument("project_path", type=str, help=project_path_help)

    _create = project.add_parser("create", help="Create new project")
    _create.set_defaults(func=async_cmd(create))
    _create.add_argument("project_path", type=str, help=project_path_help)
    _create.add_argument("--display-name", type=str)
    _create.add_argument("--public", type=str2bool, nargs="?", const=True, default=None)
    _create.add_argument("--description", type=str)
    _create.add_argument("--site-url", type=str)
    _create.add_argument("--photo-url", type=str)

    _update = project.add_parser("update", help="Update existing project")
    _update.set_defaults(func=async_cmd(update))
    _update.add_argument("project_path", type=str, help=project_path_help)
    _update.add_argument("--display-name", type=str)
    _update.add_argument("--public", type=str2bool, nargs="?", const=True, default=None)
    _update.add_argument("--description", type=str)
    _update.add_argument("--site-url", type=str)
    _update.add_argument("--photo-url", type=str)

    _delete = project.add_parser("delete", help="Delete an empty project")
    _delete.set_defaults(func=async_cmd(delete))
    _delete.add_argument("project_path", type=str, help=project_path_help)

    _transfer = project.add_parser(
        "transfer",
        help="Transfer project to another user or organization",
    )
    _transfer.set_defaults(func=async_cmd(transfer_organization))
    _transfer.add_argument("project_path", type=str, help=project_path_help)
    _transfer.add_argument("new_organization", type=str)

    _update_member_permissions = project.add_parser(
        "update-permissions",
        help="Add user to project or update user permissions",
    )
    _update_member_permissions.set_defaults(func=async_cmd(update_member_permissions))
    _update_member_permissions.add_argument("project_path", type=str, help=project_path_help)
    _update_member_permissions.add_argument("username", type=str)
    _update_member_permissions.add_argument(
        "--view",
        type=str2bool,
        nargs="?",
        const=True,
        default=None,
    )
    _update_member_permissions.add_argument(
        "--create",
        type=str2bool,
        nargs="?",
        const=True,
        default=None,
    )
    _update_member_permissions.add_argument(
        "--admin",
        type=str2bool,
        nargs="?",
        const=True,
        default=None,
    )


async def list_for_me(args):
    client = Client()
    me = await client.admin.organizations.find_me()
    result = await client.admin.projects.list_for_user(
        user_id=me["personalUserID"],
    )
    for project in result:
        org = pretty_entity_name(project["organization"]["name"])
        proj = pretty_entity_name(project["name"])
        print(f"{org}/{proj}")


async def show(args):
    client = Client()
    pq = ProjectQualifier.from_path(args.project_path)
    result = await client.admin.projects.find_by_organization_and_name(
        organization_name=pq.organization,
        project_name=pq.project,
    )
    pretty_print_graphql_result(result)


async def create(args):
    client = Client()
    pq = ProjectQualifier.from_path(args.project_path)
    organization = await client.admin.organizations.find_by_name(pq.organization)
    result = await client.admin.projects.create(
        organization_id=organization["organizationID"],
        project_name=pq.project,
        display_name=args.display_name,
        public=args.public,
        description=args.description,
        site_url=args.site_url,
        photo_url=args.photo_url,
    )
    pretty_print_graphql_result(result)


async def update(args):
    client = Client()
    pq = ProjectQualifier.from_path(args.project_path)
    project = await client.admin.projects.find_by_organization_and_name(pq.organization, pq.project)
    result = await client.admin.projects.update(
        project_id=project["projectID"],
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
    result = await client.admin.projects.delete(project_id=project["projectID"])
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
    result = await client.admin.projects.get_member_permissions(project_id=project["projectID"])
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
