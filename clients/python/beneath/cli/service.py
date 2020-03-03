from beneath.client import Client
from beneath.cli.utils import async_cmd, mb_to_bytes, parse_names, pretty_print_graphql_result, str2bool


def add_subparser(root):
  service = root.add_parser('service').add_subparsers()

  _list = service.add_parser('list')
  _list.set_defaults(func=async_cmd(show_list))
  _list.add_argument('organization', type=str)

  _create = service.add_parser('create')
  _create.set_defaults(func=async_cmd(create))
  _create.add_argument('name', type=str)
  _create.add_argument('-o', '--organization', type=str)
  _create.add_argument('--read-quota-mb', type=int, required=True)
  _create.add_argument('--write-quota-mb', type=int, required=True)

  _update = service.add_parser('update')
  _update.set_defaults(func=async_cmd(update))
  _update.add_argument('name', type=str)
  _update.add_argument('-o', '--organization', type=str)
  _update.add_argument('--new-name', type=str)
  _update.add_argument('--read-quota-mb', type=int)
  _update.add_argument('--write-quota-mb', type=int)

  _migrate = service.add_parser('migrate')
  _migrate.set_defaults(func=async_cmd(update_organization))
  _migrate.add_argument('name', type=str)
  _migrate.add_argument('-o', '--organization', type=str, required=True)
  _migrate.add_argument('--new-organization', type=str, required=True)

  _delete = service.add_parser('delete')
  _delete.set_defaults(func=async_cmd(delete))
  _delete.add_argument('name', type=str)
  _delete.add_argument('-o', '--organization', type=str)

  _add_perm = service.add_parser('update-permission')
  _add_perm.set_defaults(func=async_cmd(update_permission))
  _add_perm.add_argument('service_name', type=str)
  _add_perm.add_argument('-o', '--service-organization', type=str)
  _add_perm.add_argument('stream_name', type=str)
  _add_perm.add_argument('-p', '--stream-project', type=str)
  _add_perm.add_argument('--read', type=str2bool, nargs='?', const=True, default=None)
  _add_perm.add_argument('--write', type=str2bool, nargs='?', const=True, default=None)

  _issue_secret = service.add_parser('issue-secret')
  _issue_secret.set_defaults(func=async_cmd(issue_secret))
  _issue_secret.add_argument('service_name', type=str)
  _issue_secret.add_argument('-o', '--organization', type=str)
  _issue_secret.add_argument('--description', type=str)

  _list_secrets = service.add_parser('list-secrets')
  _list_secrets.set_defaults(func=async_cmd(list_secrets))
  _list_secrets.add_argument('service_name', type=str)
  _list_secrets.add_argument('-o', '--organization', type=str)

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
  name, org_name = parse_names(args.name, args.organization, "organization")
  org = await client.admin.organizations.find_by_name(name=org_name)
  result = await client.admin.services.create(
    name=name,
    organization_id=org['organizationID'],
    read_quota_bytes=mb_to_bytes(args.read_quota_mb),
    write_quota_bytes=mb_to_bytes(args.write_quota_mb),
  )
  pretty_print_graphql_result(result)


async def update(args):
  client = Client()
  name, org_name = parse_names(args.name, args.organization, "organization")
  service = await client.admin.services.find_by_organization_and_name(
    organization_name=org_name,
    name=name,
  )
  result = await client.admin.services.update_details(
    service_id=service['serviceID'],
    name=args.new_name,
    read_quota_bytes=mb_to_bytes(args.read_quota_mb) if args.read_quota_mb is not None else None,
    write_quota_bytes=mb_to_bytes(args.write_quota_mb) if args.write_quota_mb is not None else None,
  )
  pretty_print_graphql_result(result)


async def update_organization(args):
  client = Client()
  name, org_name = parse_names(args.name, args.organization, "organization")
  service = await client.admin.services.find_by_organization_and_name(
    organization_name=org_name,
    name=name,
  )
  new_org = await client.admin.organizations.find_by_name(args.new_organization)
  result = await client.admin.services.update_organization(
    service_id=service['serviceID'],
    organization_id=new_org['organizationID'],
  )
  pretty_print_graphql_result(result)


async def delete(args):
  client = Client()
  name, org_name = parse_names(args.name, args.organization, "organization")
  service = await client.admin.services.find_by_organization_and_name(
    organization_name=org_name,
    name=name,
  )
  result = await client.admin.services.delete(service_id=service['serviceID'])
  pretty_print_graphql_result(result)


async def update_permission(args):
  client = Client()
  service_name, org_name = parse_names(args.service_name, args.service_organization, "organization")
  stream_name, project_name = parse_names(args.stream_name, args.stream_project, "project")
  service = await client.admin.services.find_by_organization_and_name(
    organization_name=org_name,
    name=service_name,
  )
  stream = await client.admin.streams.find_by_project_and_name(
    project_name=project_name, stream_name=stream_name
  )
  result = await client.admin.services.update_permissions_for_stream(
    service_id=service['serviceID'],
    stream_id=stream['streamID'],
    read=args.read,
    write=args.write,
  )
  pretty_print_graphql_result(result)


async def issue_secret(args):
  client = Client()
  name, org_name = parse_names(args.service_name, args.organization, "organization")
  service = await client.admin.services.find_by_organization_and_name(
    organization_name=org_name,
    name=name,
  )
  result = await client.admin.services.issue_secret(
    service_id=service['serviceID'],
    description=args.description if args.description is not None else "Command-line issued secret",
  )
  print("\n" + "Keep your secret string safe. You won't be shown it again." + "\n")
  pretty_print_graphql_result(result)


async def list_secrets(args):
  client = Client()
  name, org_name = parse_names(args.service_name, args.organization, "organization")
  service = await client.admin.services.find_by_organization_and_name(
    organization_name=org_name,
    name=name,
  )
  result = await client.admin.services.list_secrets(service_id=service['serviceID'])
  pretty_print_graphql_result(result)


async def revoke_secret(args):
  client = Client()
  result = await client.admin.secrets.revoke(secret_id=args.secret_id)
  pretty_print_graphql_result(result)
