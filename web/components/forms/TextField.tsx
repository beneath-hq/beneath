import { InputBase, makeStyles, Theme } from "@material-ui/core";
import { FC } from "react";

import FormControl, { FormControlProps } from "./FormControl";

const useStyles = makeStyles((theme: Theme) => ({
  input: {
    borderRadius: "4px",
    position: "relative",
    backgroundColor: theme.palette.background.paper,
    border: "1px solid",
    borderColor: "rgba(35, 48, 70, 1)",
    padding: "10px 12px",
    marginTop: "0.3rem",
    "&:focus": {
      boxShadow: `0 0 0 2px ${theme.palette.primary.main}`,
    },
  },
  multiline: {
    padding: "0",
  },
}));

export interface TextFieldProps extends FormControlProps {
  value?: string;
  multiline?: boolean;
  rows?: number;
  rowsMax?: number;
  onChange?: React.ChangeEventHandler<HTMLTextAreaElement | HTMLInputElement>;
  onBlur?: React.FocusEventHandler<HTMLInputElement | HTMLTextAreaElement>;
}

const TextField: FC<TextFieldProps> = (props) => {
  const {
    id,
    value,
    multiline,
    rows,
    rowsMax,
    onChange,
    onBlur,
    ...others
  } = props;
  const classes = useStyles();

  return (
    <FormControl id={id} {...others}>
      <InputBase
        classes={{ input: classes.input, multiline: classes.multiline }}
        id={id}
        value={value}
        multiline={multiline}
        rows={rows}
        rowsMax={rowsMax}
        onChange={onChange}
        onBlur={onBlur}
      />
    </FormControl>
  );
};

export default TextField;
