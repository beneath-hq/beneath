import { FC } from "react";
import { Form as FormikForm, FormikFormProps } from "formik";
import { Typography } from "@material-ui/core";

export interface FormProps extends FormikFormProps {
  title?: string;
  helperText?: string | JSX.Element;
  variant?: "embedded" | "page";
  children?: any;
}

export const Form: FC<FormProps> = ({title, helperText, variant, children, ...props}) => {
  variant = variant ?? "page";
  return (
    <FormikForm>
      {title && (
        <Typography
          component={variant === "page" ? "h2" : "h2"}
          variant={variant === "page" ? "h1" : "h2"}
          gutterBottom
        >
          {title}
        </Typography>
      )}
      {helperText && (
        <Typography variant="body2" color="textSecondary">
          {helperText}
        </Typography>
      )}
      {children}
    </FormikForm>
  );
};

export default Form;
