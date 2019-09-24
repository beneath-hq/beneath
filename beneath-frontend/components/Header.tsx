import { NextRouter, withRouter } from "next/router";
import React, { FC } from "react";

import { Me } from "../apollo/types/Me";
import withMe from "../hocs/withMe";
import NextMuiLink from "./NextMuiLink";

import {
  AppBar,
  Button,
  IconButton,
  Link,
  makeStyles,
  Menu,
  MenuItem,
  Tab,
  Tabs,
  Toolbar,
} from "@material-ui/core";

import {
  Menu as MenuIcon,
  Person,
} from "@material-ui/icons";

const tabs = [
  { label: "Explore", href: "/explore", selectRegex: "^/(explore|project|stream|user).*$", external: false },
  { label: "Docs", href: "https://about.beneath.network/docs", selectRegex: "^/docs$", external: true },
  { label: "Blog", href: "https://about.beneath.network/blog", selectRegex: "^/blog$", external: true },
];

const useStyles = makeStyles((theme) => ({
  grow: {
    flexGrow: 1,
  },
  drawerButton: {
    marginRight: theme.spacing(2),
    marginLeft: 0,
    [theme.breakpoints.up("sm")]: {
      display: "none",
    },
  },
}));

interface HeaderProps extends Me {
  router: NextRouter;
  toggleMobileDrawer?: () => void;
}

const Header: FC<HeaderProps> = ({ me, router, toggleMobileDrawer }) => {
  // prepare profile menu
  const [menuAnchorEl, setMenuAnchorEl] = React.useState(null);
  const isMenuOpen = !!menuAnchorEl;
  const openMenu = (event: any) => setMenuAnchorEl(event.currentTarget);
  const closeMenu = () => setMenuAnchorEl(null);

  // compute selected tab
  const selectedTab = tabs.find((tab) => !!router.pathname.match(tab.selectRegex));
  const classes = useStyles();

  const makeMenuItem = (text: string, props: any) => {
    return (
      <MenuItem component={NextMuiLink} {...props}>{text}</MenuItem>
    );
  };

  const makeButton = (text: string, props: any) => {
    return (
      <Button component={NextMuiLink} {...props}>{text}</Button>
    );
  };

  return (
    <AppBar position="sticky">
      <Toolbar variant="dense">
        {toggleMobileDrawer && (
          <IconButton
            color="inherit"
            edge="start"
            aria-label="Open drawer"
            className={classes.drawerButton}
            onClick={toggleMobileDrawer}
          >
            <MenuIcon />
          </IconButton>
        )}
        <Link component={NextMuiLink} href="/" variant="h6" color="inherit" underline="none" noWrap>
          BENEATH
        </Link>
        <div className={classes.grow} />
        <Tabs indicatorColor="primary" textColor="primary" value={selectedTab ? selectedTab.href : false}>
          {tabs.map(({ href, label, external }) => (
            <Tab key={href} label={label} value={href} component={external ? "a" : NextMuiLink} href={href} />
          ))}
        </Tabs>
        {/* Login-specific stuff */}
        <React.Fragment>
          {!me && makeButton("Login", {
            color: "inherit",
            size: "small",
            href: "/auth",
          })}
          {me && (
            <React.Fragment>
              <IconButton edge="end" aria-haspopup="true" onClick={openMenu} color="inherit">
                <Person />
              </IconButton>
              <Menu
                anchorEl={menuAnchorEl}
                open={isMenuOpen}
                onClose={closeMenu}
                anchorOrigin={{ vertical: "top", horizontal: "right" }}
                transformOrigin={{ vertical: "top", horizontal: "right" }}
              >
                {makeMenuItem("Profile", {
                  onClick: closeMenu,
                  as: `/users/${me.user.username}`,
                  href: `/user?name=${me.user.username}`,
                })}
                {makeMenuItem("Logout", {
                  href: `/auth/logout`,
                })}
              </Menu>
            </React.Fragment>
          )}
        </React.Fragment>
      </Toolbar>
    </AppBar>
  );
};

export default withMe(withRouter(Header));
