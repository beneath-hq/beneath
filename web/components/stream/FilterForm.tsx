import avro from "avsc";
import _ from "lodash";
import {
  Button,
  Dialog,
  DialogActions,
  DialogContent,
  DialogContentText,
  DialogTitle,
  Grid,
  makeStyles,
  Theme,
} from "@material-ui/core";
import { FC, useState, useEffect } from "react";

import FilterField, { Operator, Field, FieldFilter } from "./FilterField";
import { Column, deserializeValue } from "./schema";
import CodePaper from "components/CodePaper";
import clsx from "clsx";

const useStyles = makeStyles((theme: Theme) => ({
  height: {
    height: 28,
  },
  safariButtonFix: {
    whiteSpace: "nowrap",
  },
}));

// the form
interface FilterFormProps {
  filter: any;
  index: Column[];
  onChange: (filter: any) => void;
}

const FilterForm: FC<FilterFormProps> = ({ filter, index, onChange }) => {
  const classes = useStyles();
  const [fields, setFields] = useState<Field[]>(() => {
    // filter is not empty; initialize fields with the keys present in the filter
    if (!_.isEmpty(filter)) {
      const keys = Object.keys(filter);
      const fields: Field[] = [];
      for (const key of keys) {
        const col = index.find((col) => col.name === key) as Column;
        const field = {
          name: col.name,
          type: col.inputType,
          description: col.doc,
        };
        fields.push(field);
      }
      return fields;
    }

    // filter is empty; initialize fields with the first key in the index
    const col = index[0];
    return [
      {
        name: col.name,
        type: col.inputType,
        description: col.doc,
      },
    ];
  });

  const [showAdd, setShowAdd] = useState(false);
  const [showFilter, setShowFilter] = useState(false);

  // assess whether the Add button should be shown
  useEffect(() => {
    if (fields.length < index.length) {
      // Look at previousOp. If =, then show the Add button.
      try {
        const previous = fields[fields.length - 1];
        const previousFilter = filter[previous.name];
        const previousOp = Object.keys(previousFilter)[0];
        if (previousOp === "_eq") {
          setShowAdd(true);
          return;
        }
      } catch {}
    }
    setShowAdd(false);
  }, [JSON.stringify(filter), fields]);

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

    const field = {
      name: col.name,
      type: col.inputType,
      description: col.doc,
    };
    setFields([...fields, field]);
  };

  if (fields.length === 0 || fields.length < Object.keys(filter).length) {
    addField();
  }

  const removeField = (idx: number) => {
    // nothing to remove
    if (idx + 1 > fields.length) {
      return;
    }
    for (let i = idx; i < fields.length; i++) {
      const field = fields[i];
      if (!filter[field.name]) {
        break;
      }
      delete filter[field.name];
      onChange(filter);
    }
    setFields(fields.slice(0, idx));
  };

  const onBlur = (field: Field, fieldFilter: FieldFilter) => {
    // delete all the keys from fieldFilter if there is no associated value
    for (const op in fieldFilter) {
      if (fieldFilter[op as Operator] === "") {
        delete fieldFilter[op as Operator];
      }
    }

    // if the fieldFilter is empty, then delete the field from the filter
    if (_.isEmpty(fieldFilter)) {
      delete filter[field.name];
    } else {
      filter[field.name] = fieldFilter;
    }
    onChange(filter);

    // if field is not the last one in the index, then remove subsequent fields
    // this ensures that, when editing a field's filter, no subsequent fields maintain their filter
    const idx = index.findIndex((col) => col.name === field.name);
    if (idx < index.length - 1) {
      removeField(idx + 1);
    }
  };

  return (
    <>
      {fields.map((field, idx) => {
        let initialOperator: Operator | undefined;
        let initialFieldValue: string | undefined;

        // if the filter already includes the field, it'll populate the component with the values
        if (Object.keys(filter).includes(field.name)) {
          initialOperator = Object.keys(filter[field.name])[0] as Operator;
          const serializedVal = filter[field.name][initialOperator];
          const avroType = index.find((col) => col.name === field.name)?.type as avro.Type;
          initialFieldValue = deserializeValue(avroType, serializedVal);
        }

        return (
          <Grid item key={idx}>
            <FilterField
              filter={filter}
              fields={[field]}
              initialField={field}
              initialOperator={initialOperator}
              initialFieldValue={initialFieldValue}
              cancellable={idx !== 0}
              onBlur={onBlur}
              onCancel={() => removeField(idx)}
            />
          </Grid>
        );
      })}
      {showAdd && (
        <Grid item>
          <Button onClick={addField} size="small" className={classes.height} variant="outlined">
            Add
          </Button>
        </Grid>
      )}
      {!_.isEmpty(filter) && (
        <Grid item>
          <Button
            onClick={() => setShowFilter(true)}
            size="small"
            className={clsx(classes.height, classes.safariButtonFix)}
            variant="outlined"
          >
            View filter
          </Button>
          <Dialog open={showFilter} onBackdropClick={() => setShowFilter(false)}>
            <DialogTitle>Filter</DialogTitle>
            <DialogContent>
              <DialogContentText>Use this filter with the Beneath SDK to fetch the subset of records</DialogContentText>
              <CodePaper language={"python"}>{`${JSON.stringify(filter)}`}</CodePaper>
            </DialogContent>
            <DialogActions>
              <Button onClick={() => setShowFilter(false)} color="primary">
                Close
              </Button>
            </DialogActions>
          </Dialog>
        </Grid>
      )}
    </>
  );
};

export default FilterForm;
