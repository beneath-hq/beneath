from beneath.connection import Connection


class Secrets:

  def __init__(self, conn: Connection):
    self.conn = conn

  def revoke(self, secret_id):
    result = self.conn.query_control(
      variables={
        'secretID': secret_id,
      },
      query="""
        mutation RevokeSecret($secretID: UUID!) {
          revokeSecret(secretID: $secretID)
        }
      """
    )
    return result['revokeSecret']
