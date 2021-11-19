package dev.beneath.client.utils;

public class TableIdentifier {
  public String organization;
  public String project;
  public String table;

  public TableIdentifier(String organization, String project, String table) {
    this.organization = organization;
    this.project = project;
    this.table = table;
  }

  public static TableIdentifier fromPath(String path) {
    String[] parts = Utils.splitResource("table", path);
    return new TableIdentifier(parts[0], parts[1], parts[2]);
  }

  @Override
  public String toString() {
    return String.format("%s/%s/table:%s", this.organization, this.project, this.table);
  }
}
