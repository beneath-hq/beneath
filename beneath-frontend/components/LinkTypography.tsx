import { Typography, Theme, makeStyles } from "@material-ui/core";
import Link, { LinkProps } from "next/link";
import clsx from "clsx";

type LinkTypographyProps = LinkProps & {
  bold?: boolean;
  children?: any;
};

const useStyles = makeStyles((theme: Theme) => ({
  typography: {
    cursor: "pointer",
    color: theme.palette.primary.main,
    "&:hover": {
      textDecoration: "underline",
    },
  },
  bold: {
    fontWeight: "bold",
  },
}));

export default ({ bold, children, ...linkProps }: LinkTypographyProps) => {
  const classes = useStyles();
  return (
    <Link {...linkProps}>
      <Typography className={clsx(classes.typography, bold && classes.bold)} variant="inherit">
        {children}
      </Typography>
    </Link>
  );
};
