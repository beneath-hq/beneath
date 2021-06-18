import _ from "lodash";
import {
  Divider,
  Icon,
  InputBase,
  makeStyles,
  MenuItem,
  Select,
  Theme,
  Tooltip,
  Typography,
  IconButton,
} from "@material-ui/core";
import CloseIcon from "@material-ui/icons/Close";
import WarningIcon from "@material-ui/icons/Error";
import clsx from "clsx";
import { FC, useState, useEffect } from "react";

import { InputType, serializeValue, validateValue, getPlaceholder } from "./schema";

const useStyles = makeStyles((theme: Theme) => ({
  control: {
    display: "flex",
    flexDirection: "row",
    alignItems: "center",
    borderRadius: "4px",
    position: "relative",
    border: "1px solid",
    borderColor: "rgba(83, 87, 101, 1)",
    padding: "0 12px",
    height: 28,
  },
  focused: {
    boxShadow: `0 0 0 2px ${theme.palette.primary.main}`,
  },
  error: {
    boxShadow: `0 0 0 2px ${theme.palette.error.main}`,
  },
  label: {
    fontSize: theme.typography.body1.fontSize,
  },
  iconTooltip: {
    fontSize: theme.typography.body1.fontSize,
    marginLeft: "5px",
    marginTop: "1px",
  },
  divider: {
    height: "1.4rem",
    backgroundColor: "rgba(255, 255, 255, 0.2)",
  },
  leftDivider: {
    margin: "0",
    marginLeft: "1rem",
    marginRight: "0.5rem",
  },
  rightDivider: {
    marginRight: "1rem",
    marginLeft: "0.5rem",
  },
  select: {
    padding: "0.2rem",
    paddingRight: "0.2rem !important",
    borderRadius: "4px",
    "&:hover": {
      borderRadius: "4px",
      backgroundColor: "rgba(255, 255, 255, 0.1)",
    },
  },
  selectIcon: {
    display: "none",
  },
  selectInput: {
    textAlign: "center",
    "&:focus": {
      borderRadius: "4px",
      backgroundColor: "rgba(255, 255, 255, 0.1)",
    },
  },
  valueRoot: {
    flexGrow: 0,
    minWidth: "80px",
  },
  valueInput: {
    padding: "10px 0",
  },
  clearIcon: {
    marginLeft: "-0.2rem",
    marginRight: "0.4rem",
    marginTop: "0.1rem",
    padding: "0.2rem",
  },
}));

export type Operator = "" | "_eq" | "_gt" | "_lt" | "_lte" | "_gte" | "_prefix";
export type FieldFilter = { [key in Operator]: any };

export interface Field {
  name: string;
  type: InputType;
  description?: string;
}

export interface FilterFieldProps {
  filter: any;
  fields: Field[];
  initialField: Field;
  initialOperator?: Operator;
  initialFieldValue?: string;
  cancellable?: boolean;
  onBlur?: (field: Field, fieldFilter: FieldFilter) => void;
  onCancel?: () => void;
}

const FilterField: FC<FilterFieldProps> = ({
  fields,
  initialField,
  initialOperator,
  initialFieldValue,
  cancellable,
  onBlur,
  onCancel,
}) => {
  const classes = useStyles();
  const [focused, setFocused] = useState(false);
  const [field, setField] = useState<Field>(initialField);
  const [firstOperator, setFirstOperator] = useState<Operator>(initialOperator ? initialOperator : "_eq");
  const [firstValue, setFirstValue] = useState(initialFieldValue ? initialFieldValue : "");
  const [firstHasError, setFirstHasError] = useState(false);
  // TODO: add secondary operators in the future (e.g. greater than and less than)

  const blur = () => {
    if (onBlur) {
      if (!firstHasError) {
        const fieldFilter = {} as FieldFilter;
        if (firstValue) {
          fieldFilter[firstOperator] = serializeValue(field.type, firstValue, false);
        }
        onBlur(field, fieldFilter);
      }
    }
  };

  useEffect(() => blur(), [firstOperator]);

  // make field elem
  let fieldElem: JSX.Element;
  if (fields.length > 1) {
    fieldElem = (
      <Select
        classes={{ select: classes.select, icon: classes.selectIcon }}
        input={<InputBase classes={{ root: classes.label, input: classes.selectInput }}></InputBase>}
        value={field.name}
        onChange={(e) => {
          for (const field of fields) {
            if (field.name === e.target.value) {
              setField(field);
              setFirstOperator("_eq");
              setFirstValue("");
              setFirstHasError(false);
              break;
            }
          }
        }}
      >
        {fields.map((field) => (
          <MenuItem key={field.name} value={field.name}>
            {field.name}
          </MenuItem>
        ))}
      </Select>
    );
  } else {
    fieldElem = <Typography className={classes.label}>{field.name}</Typography>;
    if (field.description) {
      fieldElem = (
        <Tooltip title={field.description} placement="top">
          {fieldElem}
        </Tooltip>
      );
    }
  }

  return (
    <div
      className={clsx(classes.control, focused && classes.focused, firstHasError && classes.error)}
      onFocus={() => setFocused(true)}
      onBlur={() => setFocused(false)}
    >
      {cancellable && (
        <IconButton
          aria-label="Clear"
          title="Clear"
          className={classes.clearIcon}
          onClick={() => {
            if (onCancel) {
              onCancel();
            }
          }}
        >
          <CloseIcon fontSize="small" />
        </IconButton>
      )}
      {fieldElem}
      <OperatorValueField
        field={field}
        operators={getOperators(field.type) || []}
        operator={firstOperator}
        value={firstValue}
        onBlur={blur}
        onHasError={setFirstHasError}
        onOperatorChange={setFirstOperator}
        onValueChange={setFirstValue}
      />
    </div>
  );
};

export default FilterField;

interface OperatorValueField {
  field: Field;
  operators: Operator[];
  operator?: Operator;
  value: string;
  onBlur: () => void;
  onHasError: (err: boolean) => void;
  onOperatorChange: (op: Operator) => void;
  onValueChange: (value: string) => void;
}

const OperatorValueField: FC<OperatorValueField> = ({
  field,
  operators,
  operator,
  value,
  onBlur,
  onHasError,
  onOperatorChange,
  onValueChange,
}) => {
  const classes = useStyles();
  const [error, setError] = useState("");
  const [valueTouched, setValueTouched] = useState(false);
  useEffect(() => onHasError(!!(valueTouched && error)), [valueTouched, error]);
  return (
    <>
      <Divider className={clsx(classes.divider, classes.leftDivider)} orientation="vertical" />
      <Select
        classes={{ select: classes.select, icon: classes.selectIcon }}
        input={<InputBase classes={{ input: classes.selectInput }}></InputBase>}
        value={operator}
        displayEmpty
        onChange={(e) => onOperatorChange(e.target.value as Operator)}
      >
        {operators.map((op) => (
          <MenuItem key={op} value={op}>
            {getOpSymbol(op)}
          </MenuItem>
        ))}
      </Select>
      {operator && (
        <>
          <Divider className={clsx(classes.divider, classes.rightDivider)} orientation="vertical" />
          <InputBase
            classes={{ root: classes.valueRoot, input: classes.valueInput }}
            value={value}
            placeholder={getPlaceholder(field.type)}
            onChange={(e) => {
              onValueChange(e.target.value);
              const err = validateValue(field.type, e.target.value);
              setError(err || "");
            }}
            onBlur={(e) => {
              if (!error) {
                onBlur();
              }
              if (!valueTouched) {
                setValueTouched(true);
              }
            }}
            onKeyDown={(e) => {
              if (e.keyCode === 13) {
                if (!error) {
                  onBlur();
                }
                if (!valueTouched) {
                  setValueTouched(true);
                }
              }
            }}
          />
          {valueTouched && error && (
            <Tooltip title={error}>
              <Icon className={classes.iconTooltip}>
                <WarningIcon fontSize="inherit" />
              </Icon>
            </Tooltip>
          )}
        </>
      )}
    </>
  );
};

const getOperators = (type: InputType, firstOperator?: Operator): Operator[] | null => {
  if (firstOperator) {
    if (firstOperator === "_lt" || firstOperator === "_lte") {
      return ["", "_gt", "_gte"];
    } else if (firstOperator === "_gt" || firstOperator === "_gte") {
      return ["", "_lt", "_lte"];
    }
    return null;
  }

  const operators: Operator[] = ["_eq", "_lt", "_gt", "_lte", "_gte"];
  if (type === "text" || type === "hex") {
    operators.push("_prefix");
  }
  return operators;
};

const getOpSymbol = (op: Operator) => {
  switch (op) {
    case "_eq": {
      return "=";
    }
    case "_gt": {
      return ">";
    }
    case "_lt": {
      return "<";
    }
    case "_lte": {
      return "<=";
    }
    case "_gte": {
      return ">=";
    }
    case "_prefix": {
      return "prefix";
    }
  }
  console.error("unexpected op: ", op);
};
