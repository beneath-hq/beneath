class ProjectQualifier:

  def __init__(self, organization: str, project: str):
    self.organization = organization
    self.project = project

  @staticmethod
  def from_path(path: str):
    parts = path.strip("/").split("/")
    if len(parts) != 2:
      raise ValueError('path must have the format "ORGANIZATION/PROJECT"')
    return ProjectQualifier(parts[0], parts[1])


class StreamQualifier:

  def __init__(self, organization: str, project: str, stream: str):
    self.organization = organization
    self.project = project
    self.stream = stream

  @staticmethod
  def from_path(path: str):
    parts = path.strip("/").split("/")
    if len(parts) != 3:
      raise ValueError('path must have the format "ORGANIZATION/PROJECT/STREAM"')
    return StreamQualifier(parts[0], parts[1], parts[2])


class ServiceQualifier:

  def __init__(self, organization: str, service: str):
    self.organization = organization
    self.service = service

  @staticmethod
  def from_path(path: str):
    parts = path.strip("/").split("/")
    if len(parts) != 2:
      raise ValueError('path must have the format "ORGANIZATION/SERVICE"')
    return ServiceQualifier(parts[0], parts[1])
  