def pretty_entity_name(name):
    return name.replace("_", "-").lower()


def split_project(path: str):
    parts = path.strip("/").split("/")
    if len(parts) != 2:
        raise ValueError('path must have the format "ORGANIZATION/PROJECT"')
    return parts


def split_resource(kind: str, path: str):
    parts = path.strip("/").split("/")
    if len(parts) != 3:
        raise ValueError(f"path must have the format 'ORGANIZATION/PROJECT/{kind.upper()}'")
    third = parts[2]
    if ":" in third:
        subparts = third.split(":")
        if len(subparts) != 2:
            raise ValueError(f"cannot parse {kind} path component '{third}'")
        if subparts[0].lower() != kind:
            raise ValueError(f"expected {kind}, got '{third}'")
        parts[2] = subparts[1]
    return parts


class ProjectQualifier:
    def __init__(self, organization: str, project: str):
        self.organization = pretty_entity_name(organization)
        self.project = pretty_entity_name(project)

    @staticmethod
    def from_path(path: str):
        parts = split_project(path)
        return ProjectQualifier(parts[0], parts[1])

    def __repr__(self):
        return f"{self.organization}/{self.project}"

    def __hash__(self):
        return hash(repr(self))

    def __eq__(self, other):
        return self.organization == other.organization and self.project == other.project


class StreamQualifier:
    def __init__(self, organization: str, project: str, stream: str):
        self.organization = pretty_entity_name(organization)
        self.project = pretty_entity_name(project)
        self.stream = pretty_entity_name(stream)

    @staticmethod
    def from_path(path: str):
        parts = split_resource("stream", path)
        return StreamQualifier(parts[0], parts[1], parts[2])

    def __repr__(self):
        return f"{self.organization}/{self.project}/stream:{self.stream}"

    def __hash__(self):
        return hash(repr(self))

    def __eq__(self, other):
        return (
            self.organization == other.organization
            and self.project == other.project
            and self.stream == other.stream
        )


class ServiceQualifier:
    def __init__(self, organization: str, project: str, service: str):
        self.organization = pretty_entity_name(organization)
        self.project = pretty_entity_name(project)
        self.service = pretty_entity_name(service)

    @staticmethod
    def from_path(path: str):
        parts = split_resource("service", path)
        return ServiceQualifier(parts[0], parts[1], parts[2])

    def __repr__(self):
        return f"{self.organization}/{self.project}/service:{self.service}"

    def __hash__(self):
        return hash(repr(self))

    def __eq__(self, other):
        return (
            self.organization == other.organization
            and self.project == other.project
            and self.service == other.service
        )


class SubscriptionQualifier:
    def __init__(self, organization: str, project: str, subscription: str):
        self.organization = pretty_entity_name(organization)
        self.project = pretty_entity_name(project)
        self.subscription = pretty_entity_name(subscription)

    @staticmethod
    def from_path(path: str):
        parts = split_resource("subscription", path)
        return SubscriptionQualifier(parts[0], parts[1], parts[2])

    def __repr__(self):
        return f"{self.organization}/{self.project}/subscription:{self.subscription}"

    def __hash__(self):
        return hash(repr(self))

    def __eq__(self, other):
        return (
            self.organization == other.organization
            and self.project == other.project
            and self.subscription == other.subscription
        )
