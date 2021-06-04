import { CollapseProps as MuiCollapseProps, Grid, Link, makeStyles } from "@material-ui/core";
import React, { FC } from "react";
import { Collapse as MuiCollapse, Typography } from "@material-ui/core";
import { ExpandMore, ExpandLess } from "@material-ui/icons";
import clsx from "clsx";

const useStyles = makeStyles((theme) => ({
  link: {
    color: theme.palette.text.primary,
    "&:hover": {
      textDecoration: "none",
      cursor: "pointer",
    },
  },
  openLink: {
    color: theme.palette.primary.main,
  },
  title: {
    fontWeight: "bold",
  },
}));

interface Props extends MuiCollapseProps {
  title: string;
  children: any;
}

export const Collapse: FC<Props> = ({ title, children }) => {
  const classes = useStyles();
  const [isOpen, setIsOpen] = React.useState(false);

  return (
    <>
      <Link onClick={() => setIsOpen((prev) => !prev)} className={clsx(classes.link, isOpen && classes.openLink)}>
        <Grid container spacing={1}>
          <Grid item>
            <Typography className={classes.title}>{title}</Typography>
          </Grid>
          <Grid item>{isOpen ? <ExpandLess /> : <ExpandMore />}</Grid>
        </Grid>
      </Link>
      {isOpen && <>{children}</>}
      {/* <MuiCollapse in={isOpen}>{children}</MuiCollapse> */}
    </>
  );
};

export default Collapse;
