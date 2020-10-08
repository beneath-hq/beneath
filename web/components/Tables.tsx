export { TableBody, TableCell, TableHead, TableRow } from "@material-ui/core";
import { makeStyles, Table as MuiTable, TableProps as MuiTableProps } from "@material-ui/core";
import clsx from "clsx";
import { FC } from "react";


const useStyles = makeStyles((theme) => ({
  medium: {
    "& th, td": {
      ...theme.typography.body1,
    },
  },
}));

export interface TableProps extends MuiTableProps {
  textSize?: "small" | "medium";
}

export const Table: FC<TableProps> = ({ textSize, className, children, ...props }) => {
  const classes = useStyles();
  return (
    <MuiTable className={clsx(className, textSize === "medium" && classes.medium)} {...props}>
      {children}
    </MuiTable>
  );
};
