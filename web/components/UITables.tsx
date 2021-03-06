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
    color: theme.palette.primary.main,
    textDecoration: "inherit",
    "&:hover": {
      textDecoration: "underline",
    },
  },
  cellExpand: {
    width: "100%",
  },
}));

// Adds textSize option to table (similar to the material-ui "size" option, but it only changes row height)
export interface UITableProps extends MuiTableProps {
  textSize?: "small" | "medium";
}

export const UITable: FC<UITableProps> = ({ textSize, className, children, ...props }) => {
  const classes = useStyles();
  return (
    <MuiTable component="div" className={clsx(className, textSize === "medium" && classes.medium)} {...props}>
      {children}
    </MuiTable>
  );
};

// Use "div"s tags instead of normal table tags (table, thead, tbody, th, tr, td, etc.) to allow
// a whole row to be a link (normally, react throws a warning when an "<a>" descends from a "<tbody>", etc.).

export const UITableHead: FC<MuiTableHeadProps> = (props) => {
  return <MuiTableHead component="div" {...props} />;
};

export const UITableBody: FC<MuiTableBodyProps> = (props) => {
  return <MuiTableBody component="div" {...props} />;
};

export const UITableRow: FC<MuiTableRowProps> = (props) => {
  return <MuiTableRow component="div" {...props} />;
};

export interface UITableCellProps extends MuiTableCellProps {
  expand?: boolean;
}

export const UITableCell: FC<UITableCellProps> = ({ className, expand, ...props }) => {
  const classes = useStyles();
  return <MuiTableCell component="div" className={clsx(className, expand && classes.cellExpand)} {...props} />;
};

// UITableRow that's a link
export const UITableLinkRow: FC<NextLinkProps> = (props) => {
  const classes = useStyles();
  return <MuiTableRow className={classes.linkRow} component={NakedLink} hover {...props} />;
};

export const UITableLinkCell: FC<NextLinkProps & MuiTableCellProps> = ({
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
