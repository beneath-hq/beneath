def pretty_entity_name(name):
  return name.replace("_", "-").lower()


class ProjectQualifier:

  def __init__(self, organization: str, project: str):
    self.organization = pretty_entity_name(organization)
    self.project = pretty_entity_name(project)

  @staticmethod
  def from_path(path: str):
    parts = path.strip("/").split("/")
    if len(parts) != 2:
      raise ValueError('path must have the format "ORGANIZATION/PROJECT"')
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
    parts = path.strip("/").split("/")
    if len(parts) != 3:
      raise ValueError('path must have the format "ORGANIZATION/PROJECT/STREAM"')
    return StreamQualifier(parts[0], parts[1], parts[2])

  def __repr__(self):
    return f"{self.organization}/{self.project}/{self.stream}"

  def __hash__(self):
    return hash(repr(self))

  def __eq__(self, other):
    return self.organization == other.organization and self.project == other.project and self.stream == other.stream


class ServiceQualifier:

  def __init__(self, organization: str, project: str, service: str):
    self.organization = pretty_entity_name(organization)
    self.project = pretty_entity_name(project)
    self.service = pretty_entity_name(service)

  @staticmethod
  def from_path(path: str):
    parts = path.strip("/").split("/")
    if len(parts) != 3:
      raise ValueError('path must have the format "ORGANIZATION/PROJECT/SERVICE"')
    return ServiceQualifier(parts[0], parts[1], parts[2])

  def __repr__(self):
    return f"{self.organization}/{self.project}/{self.service}"

  def __hash__(self):
    return hash(repr(self))

  def __eq__(self, other):
    return self.organization == other.organization and self.project == other.project and self.service == other.service
