import Drawer from "@material-ui/core/Drawer";
import Hidden from "@material-ui/core/Hidden";
import { makeStyles } from "@material-ui/core/styles";
import React, { FC } from "react";

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
  // appBarOffset: theme.mixins.toolbar, // doesn't work for variant="dense"
  appBarOffset: {
    height: 48, // see https://github.com/mui-org/material-ui/blob/next/packages/material-ui/src/Toolbar/Toolbar.js
  },
}));

interface IProps {
  mobileOpen: boolean;
  toggleMobileOpen: (event: {}) => void;
}

const ResponsivePermanentDrawer: FC<IProps> = (props) => {
  const classes = useStyles();
  return (
    <nav className={classes.drawer}>
      <Hidden xsDown implementation="css"> {/* Desktop variant */}
        <Drawer open variant="permanent" classes={{ paper: classes.drawerPaper }}>
          <div className={classes.appBarOffset} />
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
};

export default ResponsivePermanentDrawer;
