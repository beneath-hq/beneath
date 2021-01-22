from beneath.client import Client
from beneath.cli.utils import async_cmd, mb_to_bytes, pretty_print_graphql_result, str2bool


def add_subparser(root):
    organization = root.add_parser(
        "organization",
        help="Manage users and organizations",
        description="Manage users and organizations",
    ).add_subparsers()

    _update = organization.add_parser("update", help="Update a user or organization")
    _update.set_defaults(func=async_cmd(update))
    _update.add_argument("organization", type=str)
    _update.add_argument("--new-name", type=str)
    _update.add_argument("--display-name", type=str)
    _update.add_argument("--description", type=str)
    _update.add_argument("--photo-url", type=str)

    _show = organization.add_parser("show", help="Show info about a user or organization")
    _show.set_defaults(func=async_cmd(show))
    _show.add_argument("organization", type=str)

    _member_permissions = organization.add_parser(
        "show-members",
        help="List the users in an organization",
    )
    _member_permissions.set_defaults(func=async_cmd(show_member_permissions))
    _member_permissions.add_argument("organization", type=str)

    _invite_user = organization.add_parser(
        "invite-member",
        help="Invite user to join an organization",
    )
    _invite_user.set_defaults(func=async_cmd(invite_user))
    _invite_user.add_argument("organization", type=str)
    _invite_user.add_argument("username", type=str)
    _invite_user.add_argument("--view", type=str2bool, nargs="?", const=True, default=True)
    _invite_user.add_argument("--create", type=str2bool, nargs="?", const=True, default=True)
    _invite_user.add_argument(
        "--admin",
        type=str2bool,
        nargs="?",
        const=True,
        default=False,
    )

    _accept_invite = organization.add_parser(
        "accept-invite",
        help="Accept an organization invite",
    )
    _accept_invite.set_defaults(func=async_cmd(accept_invite))
    _accept_invite.add_argument("organization", type=str)

    _update_member_quota = organization.add_parser(
        "update-member-quota",
        help="Update usage quotas for an organization member",
    )
    _update_member_quota.set_defaults(func=async_cmd(update_member_quota))
    _update_member_quota.add_argument("username", type=str)
    _update_member_quota.add_argument("--read-quota-mb", type=int, required=True)
    _update_member_quota.add_argument("--write-quota-mb", type=int, required=True)
    _update_member_quota.add_argument("--scan-quota-mb", type=int, required=True)

    _leave = organization.add_parser("leave", help="Leave your current billing organization")
    _leave.set_defaults(func=async_cmd(leave))

    _remove_member = organization.add_parser(
        "remove-member",
        help="Remove member from organization",
    )
    _remove_member.set_defaults(func=async_cmd(remove_member))
    _remove_member.add_argument("username", type=str)

    _update_member_permissions = organization.add_parser(
        "update-permissions",
        help="Update organization permissions for member",
    )
    _update_member_permissions.set_defaults(func=async_cmd(update_member_permissions))
    _update_member_permissions.add_argument("organization", type=str)
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

    _create = organization.add_parser(
        "create",
        help="Create a new multi-user organization (master only)",
    )
    _create.set_defaults(func=async_cmd(create))
    _create.add_argument("name", type=str)

    _update_org_quota = organization.add_parser(
        "update-quota",
        help="Update an organization's quota hard cap (master only)",
    )
    _update_org_quota.set_defaults(func=async_cmd(update_org_quota))
    _update_org_quota.add_argument("organization", type=str)
    _update_org_quota.add_argument("--read-quota-mb", type=int, required=True)
    _update_org_quota.add_argument("--write-quota-mb", type=int, required=True)
    _update_org_quota.add_argument("--scan-quota-mb", type=int, required=True)


async def show(args):
    client = Client()
    result = await client.admin.organizations.find_by_name(name=args.organization)
    pretty_print_graphql_result(result)


async def show_member_permissions(args):
    client = Client()
    organization = await client.admin.organizations.find_by_name(args.organization)
    result = await client.admin.organizations.get_member_permissions(
        organization_id=organization["organizationID"]
    )
    pretty_print_graphql_result(result)


async def update(args):
    client = Client()
    organization = await client.admin.organizations.find_by_name(args.organization)
    result = await client.admin.organizations.update_details(
        organization_id=organization["organizationID"],
        name=args.new_name,
        display_name=args.display_name,
        description=args.description,
        photo_url=args.photo_url,
    )
    pretty_print_graphql_result(result)


async def update_org_quota(args):
    client = Client()
    organization = await client.admin.organizations.find_by_name(args.organization)
    result = await client.admin.organizations.update_quota(
        organization_id=organization["organizationID"],
        read_quota_bytes=mb_to_bytes(args.read_quota_mb),
        write_quota_bytes=mb_to_bytes(args.write_quota_mb),
        scan_quota_bytes=mb_to_bytes(args.scan_quota_mb),
    )
    pretty_print_graphql_result(result)


async def create(args):
    client = Client()
    result = await client.admin.organizations.create(args.name)
    pretty_print_graphql_result(result)


async def invite_user(args):
    client = Client()
    user = await client.admin.organizations.find_by_name(args.username)
    organization = await client.admin.organizations.find_by_name(args.organization)
    result = await client.admin.organizations.invite_user(
        organization_id=organization["organizationID"],
        user_id=user["personalUserID"],
        view=args.view,
        create=args.create,
        admin=args.admin,
    )
    pretty_print_graphql_result(result)


async def accept_invite(args):
    client = Client()
    organization = await client.admin.organizations.find_by_name(args.organization)
    result = await client.admin.organizations.accept_invite(
        organization_id=organization["organizationID"]
    )
    pretty_print_graphql_result(result)


async def update_member_quota(args):
    client = Client()
    user = await client.admin.organizations.find_by_name(args.username)
    result = await client.admin.organizations.update_user_quota(
        user_id=user["personalUserID"],
        read_quota_bytes=mb_to_bytes(args.read_quota_mb) if args.read_quota_mb >= 0 else None,
        write_quota_bytes=mb_to_bytes(args.write_quota_mb) if args.write_quota_mb >= 0 else None,
        scan_quota_bytes=mb_to_bytes(args.scan_quota_mb) if args.scan_quota_mb >= 0 else None,
    )
    pretty_print_graphql_result(result)


async def remove_user(args):
    client = Client()
    user = await client.admin.organizations.find_by_name(args.username)
    organization = await client.admin.organizations.find_by_name(args.organization)
    result = await client.admin.organizations.remove_user(
        organization_id=organization["organizationID"],
        user_id=user["personalUserID"],
    )
    pretty_print_graphql_result(result)


async def leave(args):
    client = Client()
    user = await client.admin.organizations.find_me()
    result = await client.admin.organizations.leave(user["personalUserID"])
    pretty_print_graphql_result(result)


async def remove_member(args):
    client = Client()
    user = await client.admin.organizations.find_by_name(args.username)
    result = await client.admin.organizations.leave(user["personalUserID"])
    pretty_print_graphql_result(result)


async def update_member_permissions(args):
    client = Client()
    user = await client.admin.organizations.find_by_name(args.username)
    organization = await client.admin.organizations.find_by_name(args.organization)
    result = await client.admin.users.update_permissions_for_organization(
        organization_id=organization["organizationID"],
        user_id=user["personalUserID"],
        view=args.view,
        create=args.create,
        admin=args.admin,
    )
    pretty_print_graphql_result(result)
