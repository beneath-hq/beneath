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

const useStyles = makeStyles((theme: Theme) => ({
  control: {
    display: "flex",
    flexDirection: "row",
    flexWrap: "wrap",
    alignItems: "center",
    marginTop: "0.3rem",
    borderRadius: "4px",
    position: "relative",
    backgroundColor: theme.palette.background.paper,
    border: "1px solid",
    borderColor: "rgba(35, 48, 70, 1)",
    padding: "0 12px",
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
    flexGrow: 1,
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

export type Operator = "=" | ">" | "<" | "<=" | ">=" | "prefix";
export type FieldType = "text" | "hex" | "integer" | "float" | "datetime";

export interface Field {
  name: string;
  type: FieldType;
  operators: Operator[];
  description?: string;
}

export interface FilterFieldProps {
  fields: Field[];
  cancellable?: boolean;
  onBlur?: (field: Field, op: Operator, val: string) => void;
  onCancel?: () => void;
}

const getPlaceholder = (type: FieldType) => {
  if (type === "text") {
    return "Abcd...";
  } else if (type === "hex") {
    return "0x12ab...";
  } else if (type === "integer") {
    return "1234...";
  } else if (type === "float") {
    return "1.234...";
  } else if (type === "datetime") {
    return "2006-01-02T15:04:05";
  }
  return "";
};

const validateValue = (type: FieldType, value: string): (string | null) => {
  if (value.length === 0 || type === "text") {
    return null;
  }

  if (type === "hex") {
    if (!value.match(/^0x[0-9a-fA-F]+$/)) {
      return "Expected a hexadecimal value starting with '0x'";
    }
  } else if (type === "integer") {
    if (!value.match(/^[0-9]+$/)) {
      return "Expected an integer";
    }
  } else if (type === "float") {
    try {
      parseFloat(value);
    } catch {
      return "Expected a floating-point number";
    }
  } else if (type === "datetime") {
    const t = new Date(value);
    if (isNaN(t.valueOf())) {
      return "Expected a valid timestamp";
    }
  }
  return null;
};

const FilterField: FC<FilterFieldProps> = ({ fields, cancellable, onBlur, onCancel }) => {
  const classes = useStyles();
  const [focused, setFocused] = useState(false);
  const [field, setField] = useState(fields[0]);
  const [operator, setOperator] = useState<Operator>(field.operators[0]);
  const [value, setValue] = useState("");
  const [valueTouched, setValueTouched] = useState(false);
  const [error, setError] = useState("");

  const blur = () => {
    if (onBlur && error.length === 0) {
      onBlur(field, operator, value);
    }
  };

  useEffect(() => {
    blur();
  }, [field.name, operator]);

  useEffect(() => {
    if (valueTouched) {
      const err = validateValue(field.type, value);
      if (err) {
        setError(err);
      } else if (error.length !== 0) {
        setError("");
      }
    }
  }, [valueTouched, value]);

  // Took out because formatting is too weird
  // const valueType =
  //   field.type === "text"
  //     ? "text"
  //     : field.type === "integer" || field.type === "float"
  //     ? "number"
  //     : field.type === "datetime"
  //     ? "datetime-local"
  //     : "";

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
              setOperator(field.operators[0]);
              setValue("");
              setError("");
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
      className={clsx(classes.control, focused && classes.focused, error && classes.error)}
      onFocus={() => setFocused(true)}
      onBlur={() => {
        setFocused(false);
      }}
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
      <Divider className={clsx(classes.divider, classes.leftDivider)} orientation="vertical" />
      <Select
        classes={{ select: classes.select, icon: classes.selectIcon }}
        input={<InputBase classes={{ input: classes.selectInput }}></InputBase>}
        value={operator}
        onChange={(e) => setOperator(e.target.value as Operator)}
      >
        {field.operators.map((op) => (
          <MenuItem key={op} value={op}>
            {op}
          </MenuItem>
        ))}
      </Select>
      <Divider className={clsx(classes.divider, classes.rightDivider)} orientation="vertical" />
      <InputBase
        classes={{ root: classes.valueRoot, input: classes.valueInput }}
        value={value}
        // type={valueType}
        placeholder={getPlaceholder(field.type)}
        onChange={(e) => setValue(e.target.value)}
        onBlur={(e) => {
          blur();
          if (!valueTouched) {
            setValueTouched(true);
          }
        }}
        onKeyDown={(e) => {
          if (e.keyCode === 13) {
            blur();
          }
        }}
      />
      {error && (
        <Tooltip title={error}>
          <Icon className={classes.iconTooltip}>
            <WarningIcon fontSize="inherit" />
          </Icon>
        </Tooltip>
      )}
    </div>
  );
};

export default FilterField;
