import {
  makeStyles,
  Table as MuiTable,
  TableProps as MuiTableProps,
  TableBody as MuiTableBody,
  TableBodyProps as MuiTableBodyProps,
  TableCell as MuiTableCell,
  TableCellProps as MuiTableCellProps,
  TableHead as MuiTableHead,
  TableHeadProps as MuiTableHeadProps,
  TableRow as MuiTableRow,
  TableRowProps as MuiTableRowProps,
} from "@material-ui/core";
import clsx from "clsx";
import { FC } from "react";
import { NakedLink } from "components/Link";
import { LinkProps as NextLinkProps } from "next/link";

const useStyles = makeStyles((theme) => ({
  medium: {
    "& .MuiTableCell-root": {
      fontSize: theme.typography.body1.fontSize,
      lineHeight: theme.typography.body1.lineHeight,
    },
  },
  linkRow: {
    textDecoration: "inherit",
    "&:visited": {
      color: "inherit",
    },
  },
  linkCell: {
    display: "inline-block",
    width: "100%",
    textDecoration: "inherit",
    "&:visited": {
      color: "inherit",
    },
    "&:hover": {
      textDecoration: "underline",
    },
  },
}));

// Adds textSize option to table (similar to the material-ui "size" option, but it only changes row height)
export interface TableProps extends MuiTableProps {
  textSize?: "small" | "medium";
}

export const Table: FC<TableProps> = ({ textSize, className, children, ...props }) => {
  const classes = useStyles();
  return (
    <MuiTable component="div" className={clsx(className, textSize === "medium" && classes.medium)} {...props}>
      {children}
    </MuiTable>
  );
};

// Use "div"s tags instead of normal table tags (table, thead, tbody, th, tr, td, etc.) to allow
// a whole row to be a link (normally, react throws a warning when an "<a>" descends from a "<tbody>", etc.).

export const TableHead: FC<MuiTableHeadProps> = (props) => {
  return <MuiTableHead component="div" {...props} />;
};

export const TableBody: FC<MuiTableBodyProps> = (props) => {
  return <MuiTableBody component="div" {...props} />;
};

export const TableRow: FC<MuiTableRowProps> = (props) => {
  return <MuiTableRow component="div" {...props} />;
};

export const TableCell: FC<MuiTableCellProps> = (props) => {
  return <MuiTableCell component="div" {...props} />;
};

// TableRow that's a link
export const TableLinkRow: FC<NextLinkProps> = (props) => {
  const classes = useStyles();
  return <MuiTableRow className={classes.linkRow} component={NakedLink} hover {...props} />;
};

export const TableLinkCell: FC<NextLinkProps & MuiTableCellProps> = ({
  href,
  as,
  replace,
  scroll,
  shallow,
  passHref,
  prefetch,
  children,
  ...props
}) => {
  const classes = useStyles();
  return (
    <MuiTableCell component="div" {...props}>
      <NakedLink
        className={classes.linkCell}
        href={href}
        as={as}
        replace={replace}
        scroll={scroll}
        shallow={shallow}
        passHref={passHref}
        prefetch={prefetch}
      >
        {children}
      </NakedLink>
    </MuiTableCell>
  );
};
