from beneath.connection import Connection


class _ResourceBase:
    def __init__(self, conn: Connection, dry=False):
        self.conn = conn
        self.dry = dry

    def _before_mutation(self):
        if self.dry:
            raise Exception("Cannot run mutation on a client where dry=True")
