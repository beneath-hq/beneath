package dev.beneath.client.utils;

import com.fasterxml.jackson.core.JsonProcessingException;
import com.fasterxml.jackson.databind.ObjectMapper;

public class JsonUtils {
  private static final ObjectMapper OBJECT_MAPPER = new ObjectMapper();

  public static String serialize(Object object) {
    try {
      return OBJECT_MAPPER.writeValueAsString(object);
    } catch (JsonProcessingException e) {
      throw new RuntimeException(e);
    }
  }

  public static <T> T deserialize(String jsonString, Class<T> objectType) {
    try {
      return OBJECT_MAPPER.readValue(jsonString, objectType);
    } catch (JsonProcessingException e) {
      throw new RuntimeException(e);
    }
  }
}
