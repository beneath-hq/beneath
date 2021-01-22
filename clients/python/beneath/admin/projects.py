from beneath.connection import Connection
from beneath.utils import format_entity_name


class Projects:
    def __init__(self, conn: Connection):
        self.conn = conn

    async def find_by_organization_and_name(self, organization_name, project_name):
        result = await self.conn.query_control(
            variables={
                "organizationName": format_entity_name(organization_name),
                "projectName": format_entity_name(project_name),
            },
            query="""
                query ProjectByOrganizationAndName(
                    $organizationName: String!
                    $projectName: String!
                ) {
                    projectByOrganizationAndName(
                        organizationName: $organizationName
                        projectName: $projectName
                    ) {
                        projectID
                        name
                        displayName
                        site
                        description
                        photoURL
                        public
                        createdOn
                        updatedOn
                        streams {
                            name
                        }
                        services {
                            name
                        }
                    }
                }
            """,
        )
        return result["projectByOrganizationAndName"]

    async def list_for_user(self, user_id):
        result = await self.conn.query_control(
            variables={
                "userID": user_id,
            },
            query="""
                query ProjectsForUser(
                    $userID: UUID!
                ) {
                    projectsForUser(
                        userID: $userID
                    ) {
                        projectID
                        name
                        displayName
                        site
                        description
                        photoURL
                        public
                        organization {
                            organizationID
                            name
                        }
                        createdOn
                        updatedOn
                    }
                }
            """,
        )
        return result["projectsForUser"]

    async def create(
        self,
        organization_id,
        project_name,
        display_name=None,
        public=None,
        description=None,
        site_url=None,
        photo_url=None,
    ):
        result = await self.conn.query_control(
            variables={
                "input": {
                    "organizationID": organization_id,
                    "projectName": format_entity_name(project_name),
                    "displayName": display_name,
                    "public": public,
                    "description": description,
                    "site": site_url,
                    "photoURL": photo_url,
                },
            },
            query="""
                mutation CreateProject($input: CreateProjectInput!) {
                    createProject(input: $input) {
                        projectID
                        name
                        displayName
                        public
                        site
                        description
                        photoURL
                        createdOn
                        updatedOn
                    }
                }
            """,
        )
        return result["createProject"]

    async def update(
        self,
        project_id,
        display_name=None,
        public=None,
        description=None,
        site_url=None,
        photo_url=None,
    ):
        result = await self.conn.query_control(
            variables={
                "input": {
                    "projectID": project_id,
                    "displayName": display_name,
                    "public": public,
                    "description": description,
                    "site": site_url,
                    "photoURL": photo_url,
                },
            },
            query="""
                mutation UpdateProject($input: UpdateProjectInput!) {
                    updateProject(input: $input) {
                        projectID
                        name
                        displayName
                        public
                        site
                        description
                        photoURL
                        createdOn
                        updatedOn
                    }
                }
            """,
        )
        return result["updateProject"]

    async def get_member_permissions(self, project_id):
        result = await self.conn.query_control(
            variables={
                "projectID": project_id,
            },
            query="""
                query ProjectMembers($projectID: UUID!) {
                    projectMembers(projectID: $projectID) {
                        userID
                        name
                        displayName
                        view
                        create
                        admin
                    }
                }
            """,
        )
        return result["projectMembers"]

    async def delete(self, project_id):
        result = await self.conn.query_control(
            variables={
                "projectID": project_id,
            },
            query="""
                mutation DeleteProject($projectID: UUID!) {
                    deleteProject(projectID: $projectID)
                }
            """,
        )
        return result["deleteProject"]
