from beneath.admin.base import _ResourceBase
from beneath.utils import format_entity_name


class Organizations(_ResourceBase):
    async def find_me(self):
        result = await self.conn.query_control(
            variables={},
            query="""
                query Me {
                    me {
                        organizationID
                        name
                        displayName
                        description
                        photoURL
                        createdOn
                        projects {
                            name
                        }
                        personalUserID
                        ... on PrivateOrganization {
                            updatedOn
                            readQuota
                            writeQuota
                            scanQuota
                            readUsage
                            writeUsage
                            personalUser {
                                userID
                                email
                                createdOn
                                updatedOn
                                readQuota
                                writeQuota
                                scanQuota
                                billingOrganizationID
                            }
                        }
                    }
                }
            """,
        )
        if result is None:
            raise Exception("Cannot call get_me when authenticated with a service key")
        return result["me"]

    async def find_by_name(self, name):
        result = await self.conn.query_control(
            variables={
                "name": format_entity_name(name),
            },
            query="""
                query OrganizationByName($name: String!) {
                    organizationByName(name: $name) {
                        organizationID
                        name
                        displayName
                        description
                        photoURL
                        createdOn
                        projects {
                            name
                        }
                        personalUserID
                        ... on PrivateOrganization {
                            updatedOn
                            readQuota
                            writeQuota
                            scanQuota
                            readUsage
                            writeUsage
                            personalUser {
                                userID
                                email
                                createdOn
                                updatedOn
                                readQuota
                                writeQuota
                                scanQuota
                                billingOrganizationID
                            }
                        }
                    }
                }
            """,
        )
        return result["organizationByName"]

    async def get_member_permissions(self, organization_id):
        result = await self.conn.query_control(
            variables={
                "organizationID": organization_id,
            },
            query="""
                query OrganizationMembers($organizationID: UUID!) {
                    organizationMembers(organizationID: $organizationID) {
                        userID
                        billingOrganizationID
                        name
                        displayName
                        view
                        create
                        admin
                        readQuota
                        writeQuota
                        scanQuota
                    }
                }
            """,
        )
        return result["organizationMembers"]

    async def create(self, name):
        self._before_mutation()
        result = await self.conn.query_control(
            variables={
                "name": format_entity_name(name),
            },
            query="""
                mutation CreateOrganization($name: String!) {
                    createOrganization(name: $name) {
                        organizationID
                        name
                        displayName
                        description
                        photoURL
                        createdOn
                        updatedOn
                        personalUserID
                        readQuota
                        writeQuota
                        scanQuota
                        readUsage
                        writeUsage
                        projects {
                            name
                        }
                    }
                }
            """,
        )
        return result["createOrganization"]

    async def update_details(self, organization_id, name, display_name, description, photo_url):
        self._before_mutation()
        result = await self.conn.query_control(
            variables={
                "organizationID": organization_id,
                "name": format_entity_name(name) if name else None,
                "displayName": display_name,
                "description": description,
                "photoURL": photo_url,
            },
            query="""
                mutation UpdateOrganization(
                    $organizationID: UUID!
                    $name: String
                    $displayName: String
                    $description: String
                    $photoURL: String
                ) {
                    updateOrganization(
                        organizationID: $organizationID
                        name: $name
                        displayName: $displayName
                        description: $description
                        photoURL: $photoURL
                    ) {
                        organizationID
                        name
                        displayName
                        description
                        photoURL
                        createdOn
                        updatedOn
                        personalUserID
                        readQuota
                        writeQuota
                        scanQuota
                        readUsage
                        writeUsage
                        projects {
                            name
                        }
                    }
                }
            """,
        )
        return result["updateOrganization"]

    async def update_quota(
        self,
        organization_id,
        read_quota_bytes,
        write_quota_bytes,
        scan_quota_bytes,
    ):
        self._before_mutation()
        result = await self.conn.query_control(
            variables={
                "organizationID": organization_id,
                "readQuota": read_quota_bytes,
                "writeQuota": write_quota_bytes,
                "scanQuota": scan_quota_bytes,
            },
            query="""
                mutation UpdateOrganizationQuotas(
                    $organizationID: UUID!
                    $readQuota: Int
                    $writeQuota: Int
                    $scanQuota: Int
                ) {
                    updateOrganizationQuotas(
                        organizationID: $organizationID
                        readQuota: $readQuota
                        writeQuota: $writeQuota
                        scanQuota: $scanQuota
                    ) {
                        organizationID
                        name
                        readQuota
                        writeQuota
                        scanQuota
                    }
                }
            """,
        )
        return result["updateOrganizationQuotas"]

    async def invite_user(self, organization_id, user_id, view, create, admin):
        self._before_mutation()
        result = await self.conn.query_control(
            variables={
                "userID": user_id,
                "organizationID": organization_id,
                "view": view,
                "create": create,
                "admin": admin,
            },
            query="""
                mutation InviteUserToOrganization(
                    $userID: UUID!
                    $organizationID: UUID!
                    $view: Boolean!
                    $create: Boolean!
                    $admin: Boolean!
                ) {
                    inviteUserToOrganization(
                        userID: $userID
                        organizationID: $organizationID
                        view: $view
                        create: $create
                        admin: $admin
                    )
                }
            """,
        )
        return result["inviteUserToOrganization"]

    async def accept_invite(self, organization_id):
        self._before_mutation()
        result = await self.conn.query_control(
            variables={
                "organizationID": organization_id,
            },
            query="""
                mutation AcceptOrganizationInvite($organizationID: UUID!) {
                    acceptOrganizationInvite(organizationID: $organizationID)
                }
            """,
        )
        return result["acceptOrganizationInvite"]

    async def update_user_quota(
        self,
        user_id,
        read_quota_bytes,
        write_quota_bytes,
        scan_quota_bytes,
    ):
        self._before_mutation()
        result = await self.conn.query_control(
            variables={
                "userID": user_id,
                "readQuota": read_quota_bytes,
                "writeQuota": write_quota_bytes,
                "scanQuota": scan_quota_bytes,
            },
            query="""
                mutation UpdateUserQuotas(
                    $userID: UUID!
                    $readQuota: Int
                    $writeQuota: Int
                    $scanQuota: Int
                ) {
                    updateUserQuotas(
                        userID: $userID
                        readQuota: $readQuota
                        writeQuota: $writeQuota
                        scanQuota: $scanQuota
                    ) {
                        userID
                        email
                        readQuota
                        writeQuota
                        scanQuota
                        createdOn
                        updatedOn
                    }
                }
            """,
        )
        return result["updateUserQuotas"]

    async def leave(self, user_id):
        self._before_mutation()
        result = await self.conn.query_control(
            variables={
                "userID": user_id,
            },
            query="""
                mutation LeaveBillingOrganization($userID: UUID!) {
                    leaveBillingOrganization(userID: $userID) {
                        userID
                        email
                        updatedOn
                        billingOrganizationID
                    }
                }
            """,
        )
        return result["leaveBillingOrganization"]

    async def transfer_project(self, project_id, new_organization_id):
        self._before_mutation()
        result = await self.conn.query_control(
            variables={
                "projectID": project_id,
                "organizationID": new_organization_id,
            },
            query="""
                mutation TransferProjectToOrganization(
                    $projectID: UUID!
                    $organizationID: UUID!
                ) {
                    transferProjectToOrganization(
                        projectID: $projectID
                        organizationID: $organizationID
                    ) {
                        projectID
                        organization {
                            organizationID
                            name
                        }
                    }
                }
            """,
        )
        return result["transferProjectToOrganization"]
