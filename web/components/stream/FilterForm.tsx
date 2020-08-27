import avro from "avsc";
import { FC, useState, useEffect } from "react";
import FilterField, { FieldType, Operator, Field, Filter } from "./FilterField";
import _ from "lodash";
import { Button } from "@material-ui/core";

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
  const [fields, setFields] = useState<Field[]>([]);
  const [showAdd, setShowAdd] = useState(false);

  // trigger filter update
  useEffect(() => {
    // TODO
    // console.log("triggering filter update...");
    // console.log("raw filter:");
    // console.log(filter);
    // console.log("stringify filter:");
    // console.log(JSON.stringify(filter));
    const json = JSON.stringify(filter);
    // console.log(_.isEmpty(json))
    const submit = _.isEmpty(json) ? "" : json;
    // console.log(submit)
    onChange(submit);
  }, [filter]);

  // assess whether the Add button should be shown
  useEffect(() => {
    if (filter) {
      // Look at fields and index. If there are more fields in the index, then continue.
      if (fields.length === index.length) {
        setShowAdd(false);
        return;
      }

      // Look at previousOp. If =, then show the Add button.
      const previous = fields[fields.length - 1];
      const previousFilter = filter[previous.name];
      const previousOp = Object.keys(previousFilter)[0] as Operator;
      if (previousOp === "=") {
        setShowAdd(true);
        return;
      }
    }
    setShowAdd(false);
  }, [filter, fields]);

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
  };

  const onBlur = (field: Field, fieldFilter: Filter) => {
    // TODO
    // need to see if value is empty, then delete it from the filter
    // only set filter if there is a field, fieldFilter with operation, and fieldFilter with value
    console.log("onBlur...")
    console.log(fieldFilter)
    // if (fieldFilter) {
    //   console.log("yes fieldFilter")
    //   console.log(fieldFilter)
    //   filter[field.name] = fieldFilter;
    //   console.log(filter)
    // } else {
    //   console.log("no fieldFilter")
    //   delete filter[field.name];
    // }
    // console.log("setting filter...")
    // // need to set a legit filter. this will trigger onChange(JSON.stringify(filter))
    // setFilter(filter);

    // NOTES
    // let fieldFilter = filter[field.name];
    // // detect op change
    // if (fieldFilter[op] === undefined) {
    //   // clear previous op
    //   for (let op of field.operators) {
    //     delete fieldFilter[getOp(op)];
    //   }
    //   // remove any fields after this one
    //   removeField(idx + 1);
    // }
    // fieldFilter[getOp(op)] = val;
    // filter[field.name] = fieldFilter;
    // setFilter(filter);
  };

  return (
    <div>
      {fields.map((field, index) => (
        <FilterField
          key={index}
          fields={[field]}
          cancellable={index !== 0}
          onBlur={onBlur}
          onCancel={() => removeField(index)}
        />
      ))}
      {showAdd && (
        <Button onClick={addField}>Add</Button>
      )}
    </div>
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
