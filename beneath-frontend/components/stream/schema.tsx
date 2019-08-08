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
  public includeSequenceNumber: boolean;

  constructor(stream: QueryStream_stream, includeSequenceNumber: boolean) {
    this.keyFields = stream.keyFields;
    this.avroSchema = avro.Type.forSchema(JSON.parse(stream.avroSchema)) as avro.types.RecordType;
    this.columns = [];
    this.includeSequenceNumber = includeSequenceNumber;
    for (const field of this.avroSchema.fields) {
      this.columns.push(new Column(field.name, field.name, field.type));
    }
    if (includeSequenceNumber) {
      this.columns.push(new Column("@meta.sequence_number", "Time ago", "timeago"));
    }
  }

  public makeUniqueIdentifier(record: any) {
    let id = this.keyFields.reduce((prev, curr) => `${prev}-${record[curr]}`, "");
    if (this.includeSequenceNumber) {
      const seqNo = record["@meta"] && record["@meta"].sequence_number;
      id = `${id}-${seqNo || ""}`;
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
        return <Moment fromNow ago date={new Date(val / 1000)} />;
      }

      return val.toString();
    }
    return "";
  }
}
