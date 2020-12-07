from datetime import datetime

from beneath.connection import Connection
from beneath.utils import format_graphql_time


class Users:
    def __init__(self, conn: Connection):
        self.conn = conn

    async def update_permissions_for_project(self, user_id, project_id, view, create, admin):
        result = await self.conn.query_control(
            variables={
                "userID": user_id,
                "projectID": project_id,
                "view": view,
                "create": create,
                "admin": admin,
            },
            query="""
                mutation UpdateUserProjectPermissions(
                    $userID: UUID!
                    $projectID: UUID!
                    $view: Boolean
                    $create: Boolean
                    $admin: Boolean
                ) {
                    updateUserProjectPermissions(
                        userID: $userID
                        projectID: $projectID
                        view: $view
                        create: $create
                        admin: $admin
                    ) {
                        userID
                        projectID
                        view
                        create
                        admin
                    }
                }
            """,
        )
        return result["updateUserProjectPermissions"]

    async def update_permissions_for_organization(
        self, user_id, organization_id, view, create, admin
    ):
        result = await self.conn.query_control(
            variables={
                "userID": user_id,
                "organizationID": organization_id,
                "view": view,
                "create": create,
                "admin": admin,
            },
            query="""
                mutation UpdateUserOrganizationPermissions(
                    $userID: UUID!
                    $organizationID: UUID!
                    $view: Boolean
                    $create: Boolean
                    $admin: Boolean
                ) {
                    updateUserOrganizationPermissions(
                        userID: $userID
                        organizationID: $organizationID
                        view: $view
                        create: $create
                        admin: $admin
                    ) {
                        userID
                        organizationID
                        view
                        create
                        admin
                    }
                }
            """,
        )
        return result["updateUserOrganizationPermissions"]

    async def get_usage(self, user_id, period=None, from_time=None, until=None):
        today = datetime.today()
        if (period is None) or (period == "M"):
            default_time = datetime(today.year, today.month, 1)
        elif period == "H":
            default_time = datetime(today.year, today.month, today.day, today.hour)

        result = await self.conn.query_control(
            variables={
                "userID": user_id,
                "period": period if period else "M",
                "from": format_graphql_time(from_time)
                if from_time
                else format_graphql_time(default_time),
                "until": format_graphql_time(until) if until else None,
            },
            query="""
                query GetUserUsage(
                    $userID: UUID!
                    $period: String!
                    $from: Time!
                    $until: Time
                ) {
                    getUserUsage(userID: $userID, period: $period, from: $from, until: $until) {
                        entityID
                        period
                        time
                        readOps
                        readBytes
                        readRecords
                        writeOps
                        writeBytes
                        writeRecords
                    }
                }
            """,
        )
        return result["getUserUsage"]
