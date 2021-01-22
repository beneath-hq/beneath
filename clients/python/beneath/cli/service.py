from beneath.client import Client
from beneath.utils import ProjectQualifier, ServiceQualifier, StreamQualifier
from beneath.cli.utils import (
    async_cmd,
    mb_to_bytes,
    pretty_print_graphql_result,
    str2bool,
    project_path_help,
    service_path_help,
    stream_path_help,
)


def add_subparser(root):
    service = root.add_parser(
        "service",
        help="Create and manage services",
        description="A service represents a non-user account with its own permissions, secrets, "
        "quotas and monitoring. They're used to access Beneath from production code.",
    ).add_subparsers()

    _list = service.add_parser("list", help="List services in a project")
    _list.set_defaults(func=async_cmd(show_list))
    _list.add_argument("project_path", type=str, help=project_path_help)

    _create = service.add_parser("create", help="Create a new service")
    _create.set_defaults(func=async_cmd(create))
    _create.add_argument("service_path", type=str, help=service_path_help)
    _create.add_argument("--description", type=str)
    _create.add_argument("--source-url", type=str)
    _create.add_argument("--read-quota-mb", type=int)
    _create.add_argument("--write-quota-mb", type=int)
    _create.add_argument("--scan-quota-mb", type=int)

    _update = service.add_parser("update", help="Update a service")
    _update.set_defaults(func=async_cmd(update))
    _update.add_argument("service_path", type=str, help=service_path_help)
    _update.add_argument("--description", type=str)
    _update.add_argument("--source-url", type=str)
    _update.add_argument("--read-quota-mb", type=int)
    _update.add_argument("--write-quota-mb", type=int)
    _update.add_argument("--scan-quota-mb", type=int)

    _update_perms = service.add_parser(
        "update-permissions",
        help="Add/remove service permissions for streams",
    )
    _update_perms.set_defaults(func=async_cmd(update_permissions))
    _update_perms.add_argument("service_path", type=str, help=service_path_help)
    _update_perms.add_argument("stream_path", type=str, help=stream_path_help)
    _update_perms.add_argument("--read", type=str2bool, nargs="?", const=True, default=None)
    _update_perms.add_argument("--write", type=str2bool, nargs="?", const=True, default=None)

    _delete = service.add_parser("delete", help="Delete a service")
    _delete.set_defaults(func=async_cmd(delete))
    _delete.add_argument("service_path", type=str, help=service_path_help)

    _issue_secret = service.add_parser("issue-secret", help="Issue a new service secret")
    _issue_secret.set_defaults(func=async_cmd(issue_secret))
    _issue_secret.add_argument("service_path", type=str, help=service_path_help)
    _issue_secret.add_argument("--description", type=str)

    _list_secrets = service.add_parser("list-secrets", help="List active secrets")
    _list_secrets.set_defaults(func=async_cmd(list_secrets))
    _list_secrets.add_argument("service_path", type=str, help=service_path_help)

    _revoke_secret = service.add_parser("revoke-secret", help="Revoke a service secret")
    _revoke_secret.set_defaults(func=async_cmd(revoke_secret))
    _revoke_secret.add_argument("secret_id", type=str)


async def show_list(args):
    client = Client()
    pq = ProjectQualifier.from_path(args.project_path)
    proj = await client.admin.projects.find_by_organization_and_name(pq.organization, pq.project)
    services = proj["services"]
    if (services is None) or len(services) == 0:
        print("No services found in project")
        return
    for service in services:
        pretty_print_graphql_result(service)


async def create(args):
    client = Client()
    seq = ServiceQualifier.from_path(args.service_path)
    result = await client.admin.services.create(
        organization_name=seq.organization,
        project_name=seq.project,
        service_name=seq.service,
        description=args.description,
        source_url=args.source_url,
        read_quota_bytes=mb_to_bytes(args.read_quota_mb),
        write_quota_bytes=mb_to_bytes(args.write_quota_mb),
        scan_quota_bytes=mb_to_bytes(args.scan_quota_mb),
    )
    pretty_print_graphql_result(result)


async def update(args):
    client = Client()
    seq = ServiceQualifier.from_path(args.service_path)
    result = await client.admin.services.update(
        organization_name=seq.organization,
        project_name=seq.project,
        service_name=seq.service,
        description=args.description,
        source_url=args.source_url,
        read_quota_bytes=mb_to_bytes(args.read_quota_mb),
        write_quota_bytes=mb_to_bytes(args.write_quota_mb),
        scan_quota_bytes=mb_to_bytes(args.scan_quota_mb),
    )
    pretty_print_graphql_result(result)


async def update_permissions(args):
    client = Client()
    seq = ServiceQualifier.from_path(args.service_path)
    service = await client.admin.services.find_by_organization_project_and_name(
        organization_name=seq.organization,
        project_name=seq.project,
        service_name=seq.service,
    )
    stq = StreamQualifier.from_path(args.stream_path)
    stream = await client.admin.streams.find_by_organization_project_and_name(
        organization_name=stq.organization,
        project_name=stq.project,
        stream_name=stq.stream,
    )
    result = await client.admin.services.update_permissions_for_stream(
        service_id=service["serviceID"],
        stream_id=stream["streamID"],
        read=args.read,
        write=args.write,
    )
    pretty_print_graphql_result(result)


async def delete(args):
    client = Client()
    seq = ServiceQualifier.from_path(args.service_path)
    service = await client.admin.services.find_by_organization_project_and_name(
        organization_name=seq.organization,
        project_name=seq.project,
        service_name=seq.service,
    )
    result = await client.admin.services.delete(service_id=service["serviceID"])
    pretty_print_graphql_result(result)


async def issue_secret(args):
    client = Client()
    seq = ServiceQualifier.from_path(args.service_path)
    service = await client.admin.services.find_by_organization_project_and_name(
        organization_name=seq.organization,
        project_name=seq.project,
        service_name=seq.service,
    )
    result = await client.admin.services.issue_secret(
        service_id=service["serviceID"],
        description=args.description
        if args.description is not None
        else "Command-line issued secret",
    )
    print(
        f"Keep your secret string safe. "
        f"You won't be able to see it again.\nSecret: {result['token']}"
    )


async def list_secrets(args):
    client = Client()
    seq = ServiceQualifier.from_path(args.service_path)
    service = await client.admin.services.find_by_organization_project_and_name(
        organization_name=seq.organization,
        project_name=seq.project,
        service_name=seq.service,
    )
    result = await client.admin.services.list_secrets(service_id=service["serviceID"])
    pretty_print_graphql_result(result)


async def revoke_secret(args):
    client = Client()
    result = await client.admin.secrets.revoke_service_secret(secret_id=args.secret_id)
    pretty_print_graphql_result(result)
