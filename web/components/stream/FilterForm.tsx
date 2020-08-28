import avro from "avsc";
import { FC, useState, useEffect } from "react";
import FilterField, { FieldType, Operator, Field, Filter } from "./FilterField";
import _ from "lodash";
import { Button, Grid } from "@material-ui/core";

// from schema.tsx
interface Column {
  name: string;
  type: avro.Type;
  actualType: avro.Type;
  doc?: string;
}

// the form
interface FilterFormProps {
  index: Column[];
  onChange: (filter: string) => void;
}

const FilterForm: FC<FilterFormProps> = ({ index, onChange }) => {
  const [filter, setFilter] = useState<any>({});
  const [filterJSON, setFilterJSON] = useState("");
  const [fields, setFields] = useState<Field[]>([]);
  const [showAdd, setShowAdd] = useState(false);

  // trigger filter update
  useEffect(() => {
    const submit = filterJSON === "{}" ? "" : filterJSON;
    onChange(submit);
  }, [filterJSON]);

  // assess whether the Add button should be shown
  useEffect(() => {
    if (filterJSON !== "" && filterJSON !== "{}") {
      // Look at fields and index. If there are more fields in the index, then continue.
      if (fields.length === index.length) {
        setShowAdd(false);
        return;
      }

      // Look at previousOp. If =, then show the Add button.
      const previous = fields[fields.length - 1];
      const previousFilter = filter[previous.name];
      const previousOp = Object.keys(previousFilter)[0];
      if (previousOp === "_eq") {
        setShowAdd(true);
        return;
      }
    }
    setShowAdd(false);
  }, [filterJSON, fields]);

  const addField = () => {
    let col: Column | undefined;
    if (fields.length === 0) {
      // no previous filter, we're adding the first condition
      col = index[0];
    } else {
      // adding new condition for next field
      const previous = fields[fields.length - 1];
      let takeNext = false;
      for (const indexCol of index) {
        if (indexCol.name === previous.name) {
          takeNext = true;
        } else if (takeNext) {
          col = indexCol;
          break;
        }
      }
    }
    if (!col) {
      return;
    }

    const fieldType = getFieldType(col.actualType);
    const operators = getOperators(fieldType);
    const field = {
      name: col.name,
      description: col.doc,
      type: fieldType as FieldType,
      operators,
    };
    setFields([...fields, field]);
  };

  if (fields.length === 0) {
    addField();
  }

  const removeField = (idx: number) => {
    if (idx > fields.length) {
      return;
    }
    for (let i = idx; i < fields.length; i++) {
      const field = fields[i];
      if (!filter[field.name]) {
        break;
      }
      delete filter[field.name];
    }
    setFields(fields.slice(0, idx));
    setFilter(filter);
    setFilterJSON(JSON.stringify(filter));
  };

  const onBlur = (field: Field, fieldFilter: Filter) => {
    // map fieldFilter operators to strings
    const fieldFilterStrings = convertSymbolsToCodes(fieldFilter);

    // delete all the keys from fieldFilter if there is no associated value
    for (const op in fieldFilterStrings) {
      if (fieldFilterStrings[op] === "") {
        delete fieldFilterStrings[op];
      }
    }

    // if the fieldFilterStrings is empty, then delete the field from the filter
    if (_.isEmpty(fieldFilterStrings)) {
      delete filter[field.name];
    } else {
      filter[field.name] = fieldFilterStrings;
    }

    // if field is not the last one in the index, then remove subsequent fields
    const idx = index.findIndex(col => col.name === field.name);
    if (idx < index.length - 1) {
      removeField(idx + 1);
    }

    setFilter(filter);
    setFilterJSON(JSON.stringify(filter));
  };

  return (
    <Grid container spacing={1} alignItems="center">
      {fields.map((field, index) => (
        <Grid item key={index}>
          <FilterField
            fields={[field]}
            cancellable={index !== 0}
            onBlur={onBlur}
            onCancel={() => removeField(index)}
          />
        </Grid>
      ))}
      {showAdd && (
        <Grid item>
          <Button onClick={addField}>Add</Button>
        </Grid>
      )}
    </Grid>
  );
};

export default FilterForm;

const getFieldType = (actualType: avro.Type) => {
  if (avro.Type.isType(actualType, "logical:timestamp-millis")) {
    return "datetime";
  }
  if (avro.Type.isType(actualType, "int", "long")) {
    return "integer";
  }
  if (avro.Type.isType(actualType, "float", "double")) {
    return "float";
  }
  if (avro.Type.isType(actualType, "bytes", "fixed")) {
    return "hex";
  }
  return "text";
};

const getOperators = (type: FieldType) => {
  const operators: Operator[] = ["=", "<", ">", "<=", ">="];
  if (type === "text" || type === "hex") {
    operators.push("prefix");
  }
  return operators;
};


const convertSymbolsToCodes = (fieldFilter: Filter) => {
  return _.mapKeys(fieldFilter, (_, key) => {
    return getOp(key as Operator);
  });
};

const getOp = (op: Operator) => {
  switch (op) {
    case "=": {
      return "_eq";
    }
    case ">": {
      return "_gt";
    }
    case "<": {
      return "_lt";
    }
    case "<=": {
      return "_lte";
    }
    case ">=": {
      return "_gte";
    }
    case "prefix": {
      return "_prefix";
    }
  }
  console.error("unexpected op: ", op);
};