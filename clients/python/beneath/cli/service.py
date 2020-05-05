from beneath.client import Client
from beneath.utils import ServiceQualifier, StreamQualifier
from beneath.cli.utils import async_cmd, mb_to_bytes, parse_names, pretty_print_graphql_result, str2bool


def add_subparser(root):
  service = root.add_parser('service').add_subparsers()

  _list = service.add_parser('list')
  _list.set_defaults(func=async_cmd(show_list))
  _list.add_argument('organization', type=str)

  _create = service.add_parser('create')
  _create.set_defaults(func=async_cmd(create))
  _create.add_argument('service_path', type=str)
  _create.add_argument('--read-quota-mb', type=int, required=True)
  _create.add_argument('--write-quota-mb', type=int, required=True)

  _update = service.add_parser('rename')
  _update.set_defaults(func=async_cmd(update))
  _update.add_argument('service_path', type=str)
  _update.add_argument('--new-name', type=str)

  _update_quota = service.add_parser('update-quota')
  _update_quota.set_defaults(func=async_cmd(update_quota))
  _update_quota.add_argument('service_path', type=str)
  _update_quota.add_argument('--read-quota-mb', type=int, required=True)
  _update_quota.add_argument('--write-quota-mb', type=int, required=True)

  _update_perms = service.add_parser('update-permissions')
  _update_perms.set_defaults(func=async_cmd(update_permissions))
  _update_perms.add_argument('service_path', type=str)
  _update_perms.add_argument('stream_path', type=str)
  _update_perms.add_argument('--read', type=str2bool, nargs='?', const=True, default=None)
  _update_perms.add_argument('--write', type=str2bool, nargs='?', const=True, default=None)

  _delete = service.add_parser('delete')
  _delete.set_defaults(func=async_cmd(delete))
  _delete.add_argument('service_path', type=str)

  _transfer = service.add_parser('transfer')
  _transfer.set_defaults(func=async_cmd(transfer_organization))
  _transfer.add_argument('service_path', type=str)
  _transfer.add_argument('new_organization', type=str)

  _issue_secret = service.add_parser('issue-secret')
  _issue_secret.set_defaults(func=async_cmd(issue_secret))
  _issue_secret.add_argument('service_path', type=str)
  _issue_secret.add_argument('--description', type=str)

  _list_secrets = service.add_parser('list-secrets')
  _list_secrets.set_defaults(func=async_cmd(list_secrets))
  _list_secrets.add_argument('service_path', type=str)

  _revoke_secret = service.add_parser('revoke-secret')
  _revoke_secret.set_defaults(func=async_cmd(revoke_secret))
  _revoke_secret.add_argument('secret_id', type=str)


async def show_list(args):
  client = Client()
  org = await client.admin.organizations.find_by_name(name=args.organization)
  services = org['services']
  if (services is None) or len(services) == 0:
    print("No services found in organization")
    return
  for service in services:
    pretty_print_graphql_result(service)


async def create(args):
  client = Client()
  seq = ServiceQualifier.from_path(args.service_path)
  org = await client.admin.organizations.find_by_name(name=seq.organization)
  result = await client.admin.services.create(
    name=seq.service,
    organization_id=org['organizationID'],
    read_quota_bytes=mb_to_bytes(args.read_quota_mb),
    write_quota_bytes=mb_to_bytes(args.write_quota_mb),
  )
  pretty_print_graphql_result(result)


async def update(args):
  client = Client()
  seq = ServiceQualifier.from_path(args.service_path)
  service = await client.admin.services.find_by_organization_and_name(seq.organization, seq.service)
  result = await client.admin.services.update_details(
    service_id=service['serviceID'],
    name=args.new_name,
  )
  pretty_print_graphql_result(result)


async def update_quota(args):
  client = Client()
  seq = ServiceQualifier.from_path(args.service_path)
  service = await client.admin.services.find_by_organization_and_name(seq.organization, seq.service)
  result = await client.admin.services.update_quota(
    service_id=service['serviceID'],
    read_quota_bytes=mb_to_bytes(args.read_quota_mb) if args.read_quota_mb is not None else None,
    write_quota_bytes=mb_to_bytes(args.write_quota_mb) if args.write_quota_mb is not None else None,
  )
  pretty_print_graphql_result(result)


async def update_permissions(args):
  client = Client()
  seq = ServiceQualifier.from_path(args.service_path)
  service = await client.admin.services.find_by_organization_and_name(seq.organization, seq.service)
  stq = StreamQualifier.from_path(args.stream_path)
  stream = await client.admin.streams.find_by_organization_project_and_name(
    organization_name=stq.organization,
    project_name=stq.project,
    stream_name=stq.stream,
  )
  result = await client.admin.services.update_permissions_for_stream(
    service_id=service['serviceID'],
    stream_id=stream['streamID'],
    read=args.read,
    write=args.write,
  )
  pretty_print_graphql_result(result)


async def delete(args):
  client = Client()
  seq = ServiceQualifier.from_path(args.service_path)
  service = await client.admin.services.find_by_organization_and_name(seq.organization, seq.service)
  result = await client.admin.services.delete(service_id=service['serviceID'])
  pretty_print_graphql_result(result)


async def transfer_organization(args):
  client = Client()
  seq = ServiceQualifier.from_path(args.service_path)
  service = await client.admin.services.find_by_organization_and_name(seq.organization, seq.service)
  new_org = await client.admin.organizations.find_by_name(args.new_organization)
  result = await client.admin.organizations.transfer_service(
    service_id=service['serviceID'],
    new_organization_id=new_org['organizationID'],
  )
  pretty_print_graphql_result(result)


async def issue_secret(args):
  client = Client()
  seq = ServiceQualifier.from_path(args.service_path)
  service = await client.admin.services.find_by_organization_and_name(seq.organization, seq.service)
  result = await client.admin.services.issue_secret(
    service_id=service['serviceID'],
    description=args.description if args.description is not None else "Command-line issued secret",
  )
  print(f"Keep your secret string safe. You won't be able to see it again.\nSecret: {result['token']}")


async def list_secrets(args):
  client = Client()
  seq = ServiceQualifier.from_path(args.service_path)
  service = await client.admin.services.find_by_organization_and_name(seq.organization, seq.service)
  result = await client.admin.services.list_secrets(service_id=service['serviceID'])
  pretty_print_graphql_result(result)


async def revoke_secret(args):
  client = Client()
  result = await client.admin.secrets.revoke_service_secret(secret_id=args.secret_id)
  pretty_print_graphql_result(result)
