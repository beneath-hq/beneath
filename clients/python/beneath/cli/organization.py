from beneath.client import Client
from beneath.cli.utils import async_cmd, mb_to_bytes, pretty_print_graphql_result, str2bool


def add_subparser(root):
  organization = root.add_parser('organization').add_subparsers()

  _show = organization.add_parser('show')
  _show.set_defaults(func=async_cmd(show))
  _show.add_argument('organization', type=str)

  _member_permissions = organization.add_parser('show-members')
  _member_permissions.set_defaults(func=async_cmd(show_member_permissions))
  _member_permissions.add_argument('organization', type=str)

  _rename = organization.add_parser('rename')
  _rename.set_defaults(func=async_cmd(rename))
  _rename.add_argument('organization', type=str)
  _rename.add_argument('--new-name', type=str)

  _create = organization.add_parser('create')
  _create.set_defaults(func=async_cmd(create))
  _create.add_argument('name', type=str)

  _invite_user = organization.add_parser('invite-member')
  _invite_user.set_defaults(func=async_cmd(invite_user))
  _invite_user.add_argument('organization', type=str)
  _invite_user.add_argument('username', type=str)
  _invite_user.add_argument('--view', type=str2bool, nargs='?', const=True, default=True)
  _invite_user.add_argument('--create', type=str2bool, nargs='?', const=True, default=True)
  _invite_user.add_argument(
    '--admin',
    type=str2bool,
    nargs='?',
    const=True,
    default=False,
  )

  _accept_invite = organization.add_parser('accept-invite')
  _accept_invite.set_defaults(func=async_cmd(accept_invite))
  _accept_invite.add_argument('organization', type=str)

  _update_member_quota = organization.add_parser('update-member-quota')
  _update_member_quota.set_defaults(func=async_cmd(update_member_quota))
  _update_member_quota.add_argument('username', type=str)
  _update_member_quota.add_argument('--read-quota-mb', type=int, required=True)
  _update_member_quota.add_argument('--write-quota-mb', type=int, required=True)

  _leave = organization.add_parser('leave')
  _leave.set_defaults(func=async_cmd(leave))

  _remove_member = organization.add_parser('remove-member')
  _remove_member.set_defaults(func=async_cmd(remove_member))
  _remove_member.add_argument('username', type=str)

  _update_member_permissions = organization.add_parser('update-permissions')
  _update_member_permissions.set_defaults(func=async_cmd(update_member_permissions))
  _update_member_permissions.add_argument('organization', type=str)
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

  _update_org_quota = organization.add_parser('update-quota')
  _update_org_quota.set_defaults(func=async_cmd(update_org_quota))
  _update_org_quota.add_argument('organization', type=str)
  _update_org_quota.add_argument('--read-quota-mb', type=int, required=True)
  _update_org_quota.add_argument('--write-quota-mb', type=int, required=True)


async def show(args):
  client = Client()
  result = await client.admin.organizations.find_by_name(name=args.organization)
  pretty_print_graphql_result(result)


async def show_member_permissions(args):
  client = Client()
  organization = await client.admin.organizations.find_by_name(args.organization)
  result = await client.admin.organizations.get_member_permissions(organization_id=organization['organizationID'])
  pretty_print_graphql_result(result)


async def rename(args):
  client = Client()
  organization = await client.admin.organizations.find_by_name(args.organization)
  result = await client.admin.organizations.update_name(
    organization_id=organization['organizationID'],
    name=args.new_name,
  )
  pretty_print_graphql_result(result)


async def update_org_quota(args):
  client = Client()
  organization = await client.admin.organizations.find_by_name(args.organization)
  result = await client.admin.organizations.update_quota(
    organization_id=organization['organizationID'],
    read_quota_bytes=mb_to_bytes(args.read_quota_mb) if args.read_quota_mb is not None else None,
    write_quota_bytes=mb_to_bytes(args.write_quota_mb) if args.write_quota_mb is not None else None,
  )
  pretty_print_graphql_result(result)


async def create(args):
  client = Client()
  result = await client.admin.organizations.create(args.name)
  pretty_print_graphql_result(result)


async def invite_user(args):
  client = Client()
  user = await client.admin.users.get_by_username(args.username)
  organization = await client.admin.organizations.find_by_name(args.organization)
  result = await client.admin.organizations.invite_user(
    organization_id=organization['organizationID'],
    user_id=user["userID"],
    view=args.view,
    create=args.create,
    admin=args.admin,
  )
  pretty_print_graphql_result(result)


async def accept_invite(args):
  client = Client()
  organization = await client.admin.organizations.find_by_name(args.organization)
  result = await client.admin.organizations.accept_invite(organization_id=organization['organizationID'])
  pretty_print_graphql_result(result)


async def update_member_quota(args):
  client = Client()
  user = await client.admin.users.get_by_username(args.username)
  result = await client.admin.organizations.update_user_quota(
    user_id=user['userID'],
    read_quota_bytes=mb_to_bytes(args.read_quota_mb) if args.read_quota_mb is not None else None,
    write_quota_bytes=mb_to_bytes(args.write_quota_mb) if args.write_quota_mb is not None else None,
  )
  pretty_print_graphql_result(result)


async def remove_user(args):
  client = Client()
  user = await client.admin.users.get_by_username(args.username)
  organization = await client.admin.organizations.find_by_name(args.organization)
  result = await client.admin.organizations.remove_user(
    organization_id=organization['organizationID'],
    user_id=user['userID'],
  )
  pretty_print_graphql_result(result)


async def leave(args):
  client = Client()
  user = await client.admin.users.get_me()
  result = await client.admin.organizations.leave(user["userID"])
  pretty_print_graphql_result(result)


async def remove_member(args):
  client = Client()
  user = await client.admin.users.get_by_username(args.username)
  result = await client.admin.organizations.leave(user["userID"])
  pretty_print_graphql_result(result)


async def update_member_permissions(args):
  client = Client()
  user = await client.admin.users.get_by_username(args.username)
  organization = await client.admin.organizations.find_by_name(args.organization)
  result = await client.admin.users.update_permissions_for_organization(
    organization_id=organization['organizationID'],
    user_id=user['userID'],
    view=args.view,
    create=args.create,
    admin=args.admin,
  )
  pretty_print_graphql_result(result)
