package dev.beneath.client.utils;

import org.apache.commons.lang3.StringUtils;

public class ProjectIdentifier {
  public String organization;
  public String project;

  ProjectIdentifier(String organization, String project) {
    this.organization = organization;
    this.project = project;
  }

  public static ProjectIdentifier fromPath(String path) {
    String[] parts = splitProject(path);
    return new ProjectIdentifier(parts[0], parts[1]);
  }

  @Override
  public String toString() {
    return String.format("%s/%s", this.organization, this.project);
  }

  static private String[] splitProject(String path) {
    String[] parts = StringUtils.strip(path, "/").split("/");
    if (parts.length != 2) {
      throw new RuntimeException("path must have the format \"ORGANIZATION/PROJECT\"");
    }
    return parts;
  }
}
