import avro from "avsc";
import _ from "lodash";
import numbro from "numbro";

type TimeagoType = "timeago";
export type InputType = "text" | "hex" | "integer" | "float" | "datetime";

export interface Index {
  fields: string[];
  primary: boolean;
}

export class Schema {
  public indexes: Index[];
  public primaryIndex?: Index;
  public avroSchema: avro.types.RecordType;
  public columns: Column[];

  constructor(avroSchema: string, indexes: Index[]) {
    this.indexes = indexes;
    for (const index of indexes) {
      if (index.primary) {
        this.primaryIndex = index;
        break;
      }
    }

    this.avroSchema = avro.Type.forSchema(JSON.parse(avroSchema), {
      logicalTypes: {
        decimal: Decimal,
        uuid: UUID,
        "timestamp-millis": DateType,
      },
    }) as avro.types.RecordType;

    this.columns = this.avroSchema.fields.map((field) => {
      const key = !!this.primaryIndex?.fields.includes(field.name);
      return new Column(field.name, field.name, field.type, (field as any).doc, key);
    });
  }

  public getColumns(includeTimestamp?: boolean) {
    if (includeTimestamp) {
      const tsCol = new Column("@meta.timestamp", "Time ago", "timeago", undefined, false);
      return this.columns.concat([tsCol]);
    }
    return this.columns;
  }
}

export class Column {
  public name: string;
  public displayName: string;
  public type: avro.Type | TimeagoType;
  public inputType: InputType;
  public typeName: string;
  public typeDescription: string;
  public doc: string | undefined;
  public isKey: boolean;
  public isNullable: boolean;
  public isNumeric: boolean;
  public formatter: (val: any) => string;

  constructor(
    name: string,
    displayName: string,
    type: avro.Type | TimeagoType,
    doc: string | undefined,
    isKey: boolean
  ) {
    this.name = name;
    this.displayName = displayName;
    this.type = type;
    this.doc = doc;
    this.isKey = isKey;
    this.isNullable = false;

    // unwrap union types (i.e. nullables, since unions in Beneath are always [null, actualType])
    if (this.type !== "timeago" && avro.Type.isType(this.type, "union:unwrapped")) {
      const union = this.type as avro.types.UnwrappedUnionType;

      // assert second type is actual type
      if (union.types.length !== 2 || !avro.Type.isType(union.types[0], "null")) {
        console.error("Got corrupted union type: ", union.types);
      }

      this.type = union.types[1];
      this.isNullable = true;
    }

    // get inputType
    this.inputType = this.makeInputType(this.type);

    // compute isNumeric
    this.isNumeric = false;
    if (this.type !== "timeago") {
      const t = this.type.typeName;
      if (t === "int" || t === "long" || t === "float" || t === "double") {
        this.isNumeric = true;
      }
    }

    // compute type description
    const { name: typeName, description: typeDescription } = this.makeTypeDescription(this.type);
    this.typeName = typeName;
    this.typeDescription = typeDescription;

    // make formatter
    this.formatter = this.makeFormatter();
  }

  public formatRecord(record: any) {
    return this.formatter(_.get(record, this.name));
  }

  private makeTypeDescription = (type: avro.Type | TimeagoType): { name: string, description: string }  => {
    if (type === "timeago") {
      return { name: "Timestamp", description: "Date and time with millisecond-precision" };
    }
    if (avro.Type.isType(type, "logical:timestamp-millis")) {
      return { name: "Timestamp", description: "Date and time with millisecond-precision" };
    }
    if (avro.Type.isType(type, "logical:decimal")) {
      return { name: "Numeric", description: "Integer with up to 128 digits" };
    }
    if (avro.Type.isType(type, "logical:uuid")) {
      return { name: "UUID", description: "16-byte unique identifier" };
    }
    if (avro.Type.isType(type, "int")) {
      return { name: "Int", description: "32-bit integer" };
    }
    if (avro.Type.isType(type, "long")) {
      return { name: "Long", description: "64-bit integer" };
    }
    if (avro.Type.isType(type, "float")) {
      return { name: "Float", description: "32-bit float" };
    }
    if (avro.Type.isType(type, "double")) {
      return { name: "Double", description: "64-bit float" };
    }
    if (avro.Type.isType(type, "bytes")) {
      return { name: "Bytes", description: "Variable-length byte array" };
    }
    if (avro.Type.isType(type, "fixed")) {
      const fixed = this.type as avro.types.FixedType;
      return { name: `Bytes${fixed.size}`, description: `Fixed-length byte array of size ${fixed.size}` };
    }
    if (avro.Type.isType(type, "string")) {
      return { name: "Bytes", description: "Variable-length UTF-8 string" };
    }
    if (avro.Type.isType(type, "enum")) {
      const enumT = this.type as avro.types.EnumType;
      return { name: `${enumT.name}`, description: `Enum with options: ${enumT.symbols.join(", ")}` };
    }
    if (avro.Type.isType(type, "array")) {
      const array = this.type as avro.types.ArrayType;
      const { name, description } = this.makeTypeDescription(array.itemsType);
      return { name: `${name}[]`, description: `Array of: ${description}` };
    }
    if (avro.Type.isType(type, "record")) {
      const record = this.type as avro.types.RecordType;
      const summary = record.fields.map((field) => {
        const { name } = this.makeTypeDescription(field.type);
        return field.name + ": " + name;
      });
      return { name: record.name || "record", description: `Record with fields: ${summary}` };
    }
    console.error("Unrecognized type: ", type);
    return { name: "Unknown", description: "Type not known" };
  }

  private makeInputType = (type: avro.Type | TimeagoType) => {
    if (avro.Type.isType(type, "logical:timestamp-millis")) {
      return "datetime";
    }
    if (avro.Type.isType(type, "int", "long", "logical:decimal")) {
      return "integer";
    }
    if (avro.Type.isType(type, "float", "double")) {
      return "float";
    }
    if (avro.Type.isType(type, "bytes", "fixed")) {
      return "hex";
    }
    if (avro.Type.isType(type, "string", "enum")) {
      return "text";
    }
    if (type === "timeago") {
      // shouldn't ever need this
      return "datetime";
    }
    return "text";
  }

  private makeFormatter() {
    const nonNullFormatter = this.makeNonNullFormatter();
    return (val: any) => {
      if (val === undefined || val === null) {
        return "";
      }
      return nonNullFormatter(val);
    };
  }

  private makeNonNullFormatter() {
    if (this.type === "timeago") {
      return (val: any) => new Date(val);
    }
    if (avro.Type.isType(this.type, "logical:timestamp-millis")) {
      return (val: any) => new Date(val).toISOString().slice(0, 19);
    }
    if (avro.Type.isType(this.type, "int", "long")) {
      return (val: any) => {
        try {
          return Number(val).toLocaleString("en-US");
        } catch (e) {
          return Number(val).toLocaleString();
        }
      };
    }
    if (avro.Type.isType(this.type, "float", "double")) {
      return (val: any) => {
        // handle NaN, Infinity, and -Infinity
        if (typeof val === "string") {
          return val.toString();
        }
        return numbro(val).format("0,0.000");
      };
    }
    if (avro.Type.isType(this.type, "logical:decimal")) {
      return (val: any) => {
        try {
          // @ts-ignore
          return BigInt(val).toLocaleString("en-US");
        } catch (e) {
          return BigInt(val).toLocaleString();
        }
      };
    }
    if (avro.Type.isType(this.type, "array", "record")) {
      return (val: any) => JSON.stringify(val);
    }
    return (val: any) => val;
  }
}

class DateType extends avro.types.LogicalType {
  public _fromValue(val: any) {
    return new Date(val);
  }
  public _toValue(date: any) {
    return date instanceof Date ? +date : undefined;
  }
  public _resolve(type: avro.Type) {
    if (avro.Type.isType(type, "long", "string", "logical:timestamp-millis")) {
      return this._fromValue;
    }
  }
}

class Decimal extends avro.types.LogicalType {
  public _fromValue(val: any) {
    return val;
  }
  public _toValue(val: any) {
    return val;
  }
  public _resolve(type: avro.Type) {
    if (avro.Type.isType(type, "long", "string", "logical:decimal")) {
      return this._fromValue;
    }
  }
}

class UUID extends avro.types.LogicalType {
  public _fromValue(val: any) {
    return val;
  }
  public _toValue(val: any) {
    return val;
  }
  public _resolve(type: avro.Type) {
    if (avro.Type.isType(type, "string", "logical:uuid")) {
      return this._fromValue;
    }
  }
}
