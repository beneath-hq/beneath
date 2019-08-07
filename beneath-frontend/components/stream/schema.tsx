import TableCell from "@material-ui/core/TableCell";
import avro from "avsc";
import { QueryStream_stream } from "../../apollo/types/QueryStream";

export class Schema {
  public keyFields: string[];
  public avroSchema: avro.types.RecordType;
  public columns: Column[];
  public includeSequenceNumber: boolean;

  constructor(stream: QueryStream_stream, includeSequenceNumber: boolean) {
    this.keyFields = stream.keyFields;
    this.avroSchema = avro.Type.forSchema(JSON.parse(stream.avroSchema)) as avro.types.RecordType;
    this.columns = [];
    this.includeSequenceNumber = true;
    for (const field of this.avroSchema.fields) {
      this.columns.push(new Column(field.name, field.type));
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
  public type: avro.Type;

  constructor(name: string, type: avro.Type) {
    this.name = name;
    this.type = type;
  }

  public isNumeric(): boolean {
    const t = this.type.typeName;
    return t === "int" || t === "long" || t === "float" || t === "double";
  }

  public makeTableHeaderCell(className: string | undefined) {
    return (
      <TableCell key={this.name} className={className}>
        {this.name}
      </TableCell>
    );
  }

  public makeTableCell(record: any, className: string | undefined) {
    const align = this.isNumeric() ? "right" : "left";
    return (
      <TableCell key={this.name} className={className} align={align}>
        {this.formatValue(record[this.name])}
      </TableCell>
    );
  }

  private formatValue(val: any) {
    if (val !== undefined && val !== null) {
      return val.toString();
    }
    return "";
  }
}
