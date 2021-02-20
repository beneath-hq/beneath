from beneath.admin.base import _ResourceBase


class Secrets(_ResourceBase):
    async def revoke_service_secret(self, secret_id):
        self._before_mutation()
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
