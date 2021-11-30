package dev.beneath.client.utils;

public class SubscriptionIdentifier {
  public String organization;
  public String project;
  public String subscription;

  public SubscriptionIdentifier(String organization, String project, String subscription) {
    this.organization = Utils.prettyEntityName(organization);
    this.project = Utils.prettyEntityName(project);
    this.subscription = Utils.prettyEntityName(subscription);
  }

  public static SubscriptionIdentifier fromPath(String path) {
    String[] parts = Utils.splitResource("subscription", path);
    return new SubscriptionIdentifier(parts[0], parts[1], parts[2]);
  }

  @Override
  public String toString() {
    return String.format("%s/%s/subscription:%s", this.organization, this.project, this.subscription);
  }
}
