import { makeStyles, Theme } from "@material-ui/core";
import clsx from "clsx";
import Link, { LinkProps } from "next/link";

type LinkTypographyProps = LinkProps & {
  href: string;
  bold?: boolean;
  children?: any;
};

const useStyles = makeStyles((theme: Theme) => ({
  typography: {
    cursor: "pointer",
    color: theme.palette.primary.main,
    textDecoration: "none",
    "&:hover": {
      textDecoration: "underline",
    },
  },
  bold: {
    fontWeight: "bold",
  },
}));

export default ({ bold, children, href, ...linkProps }: LinkTypographyProps) => {
  const classes = useStyles();
  const external = href.indexOf("http") === 0;
  const a = (
    <a className={clsx(classes.typography, bold && classes.bold)} href={external ? href : undefined}>
      {children}
    </a>
  );
  if (external) {
    return a;
  }
  return (
    <Link href={href} {...linkProps}>
      {a}
    </Link>
  );
};
