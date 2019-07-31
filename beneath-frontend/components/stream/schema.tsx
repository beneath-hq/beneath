import TableCell from "@material-ui/core/TableCell";
import avro from "avsc";
import { QueryStream_stream } from "../../apollo/types/QueryStream";

export class Schema {
  public keyFields: string[];
  public avroSchema: avro.types.RecordType;
  public columns: Column[];

  constructor(stream: QueryStream_stream) {
    this.keyFields = stream.keyFields;
    this.avroSchema = avro.Type.forSchema(JSON.parse(stream.avroSchema)) as avro.types.RecordType;
    this.columns = [];
    for (const field of this.avroSchema.fields) {
      this.columns.push(new Column(field.name, field.type));
    }
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

  public makeTableHeaderCell() {
    return <TableCell key={this.name}>{this.name}</TableCell>;
  }

  public makeTableCell(record: any) {
    const align = this.isNumeric() ? "right" : "left";
    return <TableCell key={this.name} align={align}>{this.formatValue(record[this.name])}</TableCell>;
  }

  private formatValue(val: any) {
    if (val !== undefined && val !== null) {
      return val.toString();
    }
    return "";
  }
}
