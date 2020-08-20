import { FormHelperText, InputLabel, makeStyles, Theme } from "@material-ui/core";
import { FC } from "react";

const useStyles = makeStyles((theme: Theme) => ({
  helper: {
    ...theme.typography.body2,
  },
}));

export type FieldErrorProps = {
  id?: string;
  error?: boolean;
  errorText?: string;
};

const FieldError: FC<FieldErrorProps> = ({ id, error, errorText }) => {
  const classes = useStyles();
  const errorTextId = errorText && id ? `${id}-error-text` : undefined;
  if (error && errorText) {
    return (
      <FormHelperText id={errorTextId} className={classes.helper}>
        {errorText}
      </FormHelperText>
    );
  }
  return <></>;
};

export default FieldError;
