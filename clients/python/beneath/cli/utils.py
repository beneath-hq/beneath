import asyncio
import argparse
import json


def async_cmd(cmd):

  def wrapped(args):
    loop = asyncio.new_event_loop()
    asyncio.set_event_loop(loop)
    loop.run_until_complete(cmd(args))

  return wrapped


def pretty_print_graphql_result(result, fields=None):
  if fields:
    result = {k: v for k, v in result.items() if k in fields}
  pretty = json.dumps(result, indent=True)
  print(pretty)


def str2bool(v) -> bool:
  if isinstance(v, bool):
    return v
  if v.lower() in ('yes', 'true', 't', 'y', '1'):
    return True
  if v.lower() in ('no', 'false', 'f', 'n', '0'):
    return False
  raise argparse.ArgumentTypeError('Boolean value expected.')


def mb_to_bytes(v):
  return v * 1000000


def parse_names(name, explicit_group, group_type):
  names = name.split('/')
  if (len(names) > 0) and names[0] == '':
    names = names[1:]
  if len(names) > 2:
    raise argparse.ArgumentTypeError("Cannot parse identifier with more than one '/'")
  if (len(names) == 2) and (explicit_group is not None):
    raise argparse.ArgumentTypeError(
      "Cannot read {} set in both name (before the '/') and as explicit arg".format(group_type)
    )
  if len(names) == 2:
    return names[1], names[0]
  if not explicit_group:
    raise argparse.ArgumentTypeError("Must specify {}".format(group_type))
  return names[0], explicit_group
