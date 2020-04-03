import { makeStyles, Theme } from "@material-ui/core/styles";
import TextField, { TextFieldProps } from "@material-ui/core/TextField";
import Typography from "@material-ui/core/Typography";
import React, { FunctionComponent } from "react";

export type BNTextFieldProps = TextFieldProps & { errorText?: string, successText?: string | undefined };

const BNTextField: FunctionComponent<BNTextFieldProps> = ({ helperText, errorText, successText, ...props }) => {
  let elem: JSX.Element | undefined;
  if (helperText || errorText) {
    elem = (
      <>
        {helperText && <Typography variant="caption">{helperText}</Typography>}
        {helperText && errorText && (
          <>
            <br />
            <br />
          </>
        )}
        {errorText && (
          <Typography variant="caption" color="error">
            {errorText}
          </Typography>
        )}
        {(helperText || errorText) && successText && (
          <>
            <br />
            <br />
          </>
        )}
        {successText && (
          <Typography variant="caption" color="primary">
            {successText}
          </Typography>
        )}
      </>
    );
  }

  return <TextField helperText={elem} {...props} />;
};

export default BNTextField;
