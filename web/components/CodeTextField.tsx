import { makeStyles, Theme } from "@material-ui/core/styles";
import React, { FunctionComponent } from "react";

import BNTextField, { BNTextFieldProps } from "./BNTextField";

const useStyles = makeStyles((theme: Theme) => ({
}));

const handleTabInput = (event: any) => {
  if (event.keyCode === 9) { // 9 = tab key
    const start = event.target.selectionStart;
    const end = event.target.selectionEnd;

    let value = event.target.value;
    value = value.substring(0, start) + "    " + value.substring(end);
    event.target.value = value;

    event.target.selectionStart = event.target.selectionEnd = start + 4;

    event.preventDefault();
    return false;
  }
};

const CodeTextField: FunctionComponent<BNTextFieldProps> = (props) => {
  const classes = useStyles();
  return <BNTextField onKeyDown={handleTabInput} {...props} />;
};

export default CodeTextField;
