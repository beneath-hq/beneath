from beneath.connection import Connection


class Secrets:

  def __init__(self, conn: Connection):
    self.conn = conn

  async def revoke(self, secret_id):
    result = await self.conn.query_control(
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
