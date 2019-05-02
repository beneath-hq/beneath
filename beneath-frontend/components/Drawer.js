import React, { Component } from "react";
import PropTypes from "prop-types";
import Hidden from "@material-ui/core/Hidden";
import Drawer from "@material-ui/core/Drawer";
import { makeStyles } from "@material-ui/core/styles";

const drawerWidth = 240;

const useStyles = makeStyles((theme) => ({
  drawer: {
    [theme.breakpoints.up("sm")]: {
      width: drawerWidth,
      flexShrink: 0
    }
  },
  drawerPaper: {
    width: drawerWidth,
  },
}));

const ResponsivePermanentDrawer = (props) => {
  const classes = useStyles();
  return (
    <nav className={classes.drawer}>
      <Hidden xsDown implementation="css"> {/* Desktop variant */}
        <Drawer open variant="permanent" classes={{ paper: classes.drawerPaper }}>
          {props.children}
        </Drawer>
      </Hidden>
      <Hidden smUp implementation="css"> {/* Mobile variant */}
        <Drawer open={ props.mobileOpen } variant="temporary" classes={{ paper: classes.drawerPaper }} 
          ModalProps={{ keepMounted: true }} onClose={props.toggleMobileOpen}
        >
          {props.children}
        </Drawer>
      </Hidden>
    </nav>
  );
}

ResponsivePermanentDrawer.propTypes = {
  mobileOpen: PropTypes.bool,
  toggleMobileOpen: PropTypes.func,
};

export default ResponsivePermanentDrawer;
