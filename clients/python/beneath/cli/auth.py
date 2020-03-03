from beneath import config
from beneath.cli.utils import async_cmd
from beneath.client import Client
from beneath.connection import BeneathError


def add_subparser(root):
  _auth = root.add_parser('auth')
  _auth.set_defaults(func=async_cmd(auth))
  _auth.add_argument('secret', type=str)


async def auth(args):
  try:
    client = Client(secret=args.secret)
    await client.connection.ensure_connected()
    config.write_secret(args.secret)
    print("You have authenticated successfully!")
  except BeneathError:
    config.write_secret("")
    print("Your attempt to authenticate failed. Are you using an API secret generated in the Beneath web app?")
