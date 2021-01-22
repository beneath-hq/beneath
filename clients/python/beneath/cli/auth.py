from beneath import config
from beneath.cli.utils import async_cmd
from beneath.client import Client
from beneath.connection import AuthenticationError


def add_subparser(root):
    _auth = root.add_parser(
        "auth",
        help="Login to your Beneath account",
        description="For authentication instructions, visit "
        "https://about.beneath.dev/docs/quick-starts/install-sdk/",
    )
    _auth.set_defaults(func=async_cmd(auth))
    _auth.add_argument(
        "secret",
        type=str,
        help="Secret to use when making requests",
    )


async def auth(args):
    try:
        client = Client(secret=args.secret)
        await client.connection.ensure_connected()
        config.write_secret(args.secret)
        print("You have authenticated successfully!")
    except AuthenticationError:
        config.write_secret("")
        print(
            "Your attempt to authenticate failed. "
            "Are you using an API secret generated in the Beneath web app?"
        )
