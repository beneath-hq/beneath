package dev.beneath.client.utils;

public class ProjectIdentifier {
  public String organization;
  public String project;

  ProjectIdentifier(String organization, String project) {
    this.organization = organization;
    this.project = project;
  }

  @Override
  public String toString() {
    return String.format("%s/%s", this.organization, this.project);
  }
}
