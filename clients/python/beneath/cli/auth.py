import asyncio
from datetime import datetime, timedelta
import webbrowser

from beneath import config
from beneath.admin.client import AdminClient
from beneath.cli.utils import async_cmd
from beneath.client import Client
from beneath.connection import AuthenticationError, Connection


def add_subparser(root):
    _auth = root.add_parser(
        "auth",
        help="Login to your Beneath account",
        description="Opens your browser to authenticate your local environment",
    )
    _auth.set_defaults(func=async_cmd(auth))
    _auth.add_argument(
        "--secret",
        type=str,
        help="Manually issued authentication secret to use",
    )


async def auth(args):
    async def run_ticket_flow():
        admin = AdminClient(connection=Connection())
        ticket = await admin.users.create_auth_ticket("Beneath CLI")
        ticket_id = ticket["authTicketID"]

        url = f"{config.BENEATH_FRONTEND_HOST}/-/auth/ticket/{ticket_id}"
        print(f"Opening browser to authenticate...")
        print(f"If your browser does not open, go to: {url}")
        webbrowser.open(url)

        start = datetime.now()
        while True:
            await asyncio.sleep(3)

            if datetime.now() - start > timedelta(minutes=10):
                print("Authentication request timed out")
                exit(1)

            ticket = await admin.users.get_auth_ticket(ticket_id)
            if ticket is None:
                print("Authentication failed")
                exit(1)

            if type(ticket["issuedSecret"]) is dict:
                token = ticket["issuedSecret"]["token"]
                return token

    async def validate_secret(token):
        try:
            client = Client(secret=token)
            await client.connection.ensure_connected()
            config.write_secret(token)
            print("You have authenticated successfully!")
        except AuthenticationError:
            config.write_secret("")
            print("Your attempt to authenticate failed.")
            exit(1)

    token = args.secret
    if not token:
        token = await run_ticket_flow()
    await validate_secret(token)
