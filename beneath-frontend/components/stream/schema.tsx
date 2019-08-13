import avro from "avsc";
import _ from "lodash";

import TableCell from "@material-ui/core/TableCell";
import Moment from "react-moment";

import { QueryStream_stream } from "../../apollo/types/QueryStream";

type TimeagoType = "timeago";

export class Schema {
  public keyFields: string[];
  public avroSchema: avro.types.RecordType;
  public columns: Column[];
  public includeTimestamp: boolean;

  constructor(stream: QueryStream_stream, includeTimestamp: boolean) {
    this.keyFields = stream.keyFields;
    this.avroSchema = avro.Type.forSchema(JSON.parse(stream.avroSchema)) as avro.types.RecordType;
    this.columns = [];
    this.includeTimestamp = includeTimestamp;
    for (const field of this.avroSchema.fields) {
      this.columns.push(new Column(field.name, field.name, field.type));
    }
    if (includeTimestamp) {
      this.columns.push(new Column("@meta.timestamp", "Time ago", "timeago"));
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

  constructor(name: string, displayName: string, type: avro.Type | TimeagoType) {
    this.name = name;
    this.displayName = displayName;
    this.type = type;
  }

  public isNumeric(): boolean {
    if (this.type === "timeago") {
      return false;
    }
    const t = this.type.typeName;
    return t === "int" || t === "long" || t === "float" || t === "double";
  }

  public makeTableHeaderCell(className: string | undefined) {
    return (
      <TableCell key={this.name} className={className}>
        {this.displayName}
      </TableCell>
    );
  }

  public makeTableCell(record: any, className: string | undefined) {
    const align = this.isNumeric() ? "right" : "left";
    return (
      <TableCell key={this.name} className={className} align={align}>
        {this.formatValue(_.get(record, this.name))}
      </TableCell>
    );
  }

  private formatValue(val: any) {
    if (val !== undefined && val !== null) {
      if (this.type === "timeago") {
        return <Moment fromNow ago date={new Date(val)} />;
      }

      if (this.name === "day" || this.name === "time") {
        // TODODODO
        return new Date(val).toISOString().slice(0, 19);
      }

      return val.toString();
    }
    return "";
  }
}
