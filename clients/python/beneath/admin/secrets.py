from beneath.connection import Connection


class Secrets:
    def __init__(self, conn: Connection):
        self.conn = conn

    async def revoke_service_secret(self, secret_id):
        result = await self.conn.query_control(
            variables={
                "secretID": secret_id,
            },
            query="""
                mutation RevokeServiceSecret($secretID: UUID!) {
                    revokeServiceSecret(secretID: $secretID)
                }
            """,
        )
        return result["revokeServiceSecret"]
