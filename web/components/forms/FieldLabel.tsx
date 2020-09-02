import { FormHelperText, InputLabel, makeStyles, Theme } from "@material-ui/core";
import { FC } from "react";
import clsx from "clsx";

const useStyles = makeStyles((theme: Theme) => ({
  label: {
    top: "none",
    left: "none",
    transform: "none",
    position: "relative",
    ...theme.typography.h4,
    // ...theme.typography.body1,
  },
  helper: {
    ...theme.typography.body2,
  },
}));

export type FieldLabelProps = {
  id?: string;
  label?: React.ReactNode;
  helperText?: string;
  required?: boolean;
};

const FieldLabel: FC<FieldLabelProps> = ({ id, label, helperText, required }) => {
  const classes = useStyles();
  const inputLabelId = id ? `${id}-label` : undefined;
  const helperTextId = helperText && id ? `${id}-helper-text` : undefined;
  return (
    <>
      {label && (
        <InputLabel
          classes={{ formControl: clsx(classes.label, helperText) }}
          disableAnimation={true}
          shrink={false}
          id={inputLabelId}
          htmlFor={id}
        >
          {label}
          {!required && " (optional)"}
        </InputLabel>
      )}
      {helperText && (
        <FormHelperText id={helperTextId} className={classes.helper}>
          {helperText}
        </FormHelperText>
      )}
    </>
  );
};

export default FieldLabel;
