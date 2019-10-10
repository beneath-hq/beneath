import avro from "avsc";
import _ from "lodash";
import dynamic from "next/dynamic";
import numbro from "numbro";
const Moment = dynamic(import("react-moment"), { ssr: false });

import { QueryStream_stream } from "../../apollo/types/QueryStream";

type TimeagoType = "timeago";

export class Schema {
  public keyFields: string[];
  public avroSchema: avro.types.RecordType;
  public columns: Column[];
  public includeTimestamp: boolean;

  constructor(stream: QueryStream_stream, includeTimestamp: boolean) {
    this.keyFields = stream.keyFields;
    this.avroSchema = avro.Type.forSchema(JSON.parse(stream.avroSchema), {
      logicalTypes: {
        decimal: Decimal,
        "timestamp-millis": DateType,
      },
    }) as avro.types.RecordType;
    this.columns = [];
    this.includeTimestamp = includeTimestamp;
    for (const field of this.avroSchema.fields) {
      const key = this.keyFields.includes(field.name);
      this.columns.push(new Column(field.name, field.name, field.type, key, (field as any).doc));
    }
    if (includeTimestamp) {
      this.columns.push(new Column("@meta.timestamp", "Time ago", "timeago", false, undefined));
    }
  }

  public makeUniqueIdentifier(record: any) {
    let id = this.keyFields.reduce((prev, curr) => `${prev}-${record[curr]}`, "");
    if (this.includeTimestamp) {
      const ts = record["@meta"] && record["@meta"].timestamp;
      id = `${id}-${ts || ""}`;
    }
    return id;
  }
}

class Column {
  public name: string;
  public displayName: string;
  public type: avro.Type | TimeagoType;
  public key: boolean;
  public doc: string | undefined;

  constructor(name: string, displayName: string, type: avro.Type | TimeagoType, key: boolean, doc: string | undefined) {
    this.name = name;
    this.displayName = displayName;
    this.type = type;
    this.key = key;
    this.doc = doc;
  }

  public isNumeric(): boolean {
    if (this.type === "timeago") {
      return false;
    }
    const t = this.type.typeName;
    return t === "int" || t === "long" || t === "float" || t === "double";
  }

  public formatRecord(record: any) {
    return this.formatValue(_.get(record, this.name));
  }

  private formatValue(val: any) {
    if (val !== undefined && val !== null) {
      if (this.type === "timeago") {
        return <Moment fromNow ago date={new Date(val)} />;
      }

      if (avro.Type.isType(this.type, "logical:timestamp-millis")) {
        return new Date(val).toISOString().slice(0, 19);
      }

      if (avro.Type.isType(this.type, "int", "long")) {
        try {
          return Number(val).toLocaleString("en-US");
        } catch (e) {
          return Number(val).toLocaleString();
        }
      }

      if (avro.Type.isType(this.type, "float", "double")) {
        return numbro(val).format("0,0.000");
      }

      if (avro.Type.isType(this.type, "logical:decimal")) {
        try {
          // @ts-ignore
          return BigInt(val).toLocaleString("en-US");
        } catch (e) {
          return BigInt(val).toLocaleString();
        }
      }

      return val.toString();
    }
    return "";
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
