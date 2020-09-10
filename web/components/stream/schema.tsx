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

    // make formatter
    this.formatter = this.makeFormatter();
  }

  public formatRecord(record: any) {
    return this.formatter(_.get(record, this.name));
  }

  private makeInputType = (type: avro.Type | TimeagoType) => {
    if (avro.Type.isType(type, "logical:timestamp-millis")) {
      return "datetime";
    }
    if (avro.Type.isType(type, "int", "long")) {
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
