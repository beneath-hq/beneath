package dev.beneath.client.utils;

import java.nio.ByteBuffer;
import java.util.UUID;

import org.apache.commons.lang3.StringUtils;

public class Utils {
  public static String formatEntityName(String name) {
    return name.replace("-", "_").toLowerCase();
  }

  public static String prettyEntityName(String name) {
    return name.replace("_", "-").toLowerCase();
  }

  public static String[] splitResource(String kind, String path) {
    String[] parts = StringUtils.strip(path, "/").split("/");
    if (parts.length != 3) {
      throw new RuntimeException(
          String.format("path must have the format 'ORGANIZATION/PROJECT/%s", kind.toUpperCase()));
    }
    String third = parts[2];
    if (third.contains(":")) {
      String[] subparts = third.split(":");
      if (subparts.length != 2) {
        throw new RuntimeException(String.format("cannot parse %s path component '%s'", kind, third));
      }
      if (!kind.equals(subparts[0].toLowerCase())) {
        throw new RuntimeException(String.format("expected %s, got '%s'", kind, third));
      }
      parts[2] = subparts[1];
    }
    return parts;
  }

  public static byte[] uuidToBytes(UUID uuid) {
    ByteBuffer bb = ByteBuffer.wrap(new byte[16]);
    bb.putLong(uuid.getMostSignificantBits());
    bb.putLong(uuid.getLeastSignificantBits());
    return bb.array();
  }
}
