from beneath.admin.organizations import Organizations
from beneath.admin.projects import Projects
from beneath.admin.secrets import Secrets
from beneath.admin.services import Services
from beneath.admin.streams import Streams
from beneath.admin.users import Users
from beneath.connection import Connection


class AdminClient:
    """
    AdminClient isolates control-plane features.

    Args:
        connection (Connection): An authenticated connection to Beneath.
    """

    def __init__(self, connection: Connection):
        self.connection = connection
        self.organizations = Organizations(self.connection)
        self.projects = Projects(self.connection)
        self.secrets = Secrets(self.connection)
        self.services = Services(self.connection)
        self.streams = Streams(self.connection)
        self.users = Users(self.connection)
