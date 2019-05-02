import React from "react";
import Breadcrumbs from "@material-ui/core/Breadcrumbs";
import Divider from "@material-ui/core/Divider";
import Link from "@material-ui/core/Link";
import NavigateNextIcon from "@material-ui/icons/NavigateNext";
import Typography from "@material-ui/core/Typography";
import { makeStyles } from "@material-ui/core/styles";
import { withRouter } from "next/router";

const useStyles = makeStyles((theme) => ({
  breadcrumbs: {
    paddingLeft: theme.spacing(1),
  },
  content: {
  },
  divider: {
    marginTop: theme.spacing(1),
  },
}));

const Subheader = withRouter(({ router }) => {
  const classes = useStyles();
  return (
    <div className={classes.content}>
      <Breadcrumbs className={classes.breadcrumbs} separator={<NavigateNextIcon fontSize="small" />} aria-label="Breadcrumb">
        { router.pathname.split("/").slice(1).map((crumb, idx) => (
          <Typography key={idx} color="textPrimary">{crumb}</Typography>
        )) }
      </Breadcrumbs>
      <Divider className={classes.divider} />
    </div>
  );
});

export default Subheader;
