package dev.beneath;

public class DbzRecordSchema {
  private String keySchema;
  private String valueSchema;

  public DbzRecordSchema() {
  }

  public DbzRecordSchema(String keySchema, String valueSchema) {
    this.setKeySchema(keySchema);
    this.setValueSchema(valueSchema);
  }

  public String getKeySchema() {
    return keySchema;
  }

  public void setKeySchema(String keySchema) {
    this.keySchema = keySchema;
  }

  public String getValueSchema() {
    return valueSchema;
  }

  public void setValueSchema(String valueSchema) {
    this.valueSchema = valueSchema;
  }

  @Override
  public int hashCode() {
    final int prime = 31;
    int result = 1;
    result = prime * result + ((keySchema == null) ? 0 : keySchema.hashCode());
    result = prime * result + ((valueSchema == null) ? 0 : valueSchema.hashCode());
    return result;
  }

  @Override
  public boolean equals(Object obj) {
    if (this == obj)
      return true;
    if (obj == null)
      return false;
    if (getClass() != obj.getClass())
      return false;
    DbzRecordSchema other = (DbzRecordSchema) obj;
    if (keySchema == null) {
      if (other.keySchema != null)
        return false;
    } else if (!keySchema.equals(other.keySchema))
      return false;
    if (valueSchema == null) {
      if (other.valueSchema != null)
        return false;
    } else if (!valueSchema.equals(other.valueSchema))
      return false;
    return true;
  }
}
