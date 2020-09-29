import { useRouter } from "next/router";
import React, { FC } from "react";

import useMe from "../../hooks/useMe";
import BeneathLogo from "./BeneathLogo";
import UsageIndicator from "../metrics/user/UsageIndicator";
import { NakedLink } from "../Link";

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
  Typography,
} from "@material-ui/core";

import { Menu as MenuIcon, Person } from "@material-ui/icons";

const tabs = [
  { label: "Docs", href: "https://about.beneath.dev/docs", selectRegex: "^$", external: true },
  { label: "Console", href: "/", selectRegex: "^/.*$", external: false },
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
  logo: {
    height: "30px",
    marginRight: "1rem",
  },
  tabs: {
    paddingRight: theme.spacing(2),
  },
  menuPaper: {
    minWidth: "250px",
  },
  menuItemHeader: {
    borderBottom: `1px solid rgba(255, 255, 255, 0.175)`,
  },
  menuItemUsage: {
    borderBottom: `1px solid rgba(255, 255, 255, 0.175)`,
  },
}));

interface HeaderProps {
  toggleMobileDrawer?: () => void;
}

const Header: FC<HeaderProps> = ({ toggleMobileDrawer }) => {
  // prepare profile menu
  const [menuAnchorEl, setMenuAnchorEl] = React.useState(null);
  const isMenuOpen = !!menuAnchorEl;
  const openMenu = (event: any) => setMenuAnchorEl(event.currentTarget);
  const closeMenu = () => setMenuAnchorEl(null);

  const me = useMe();
  const router = useRouter();

  // compute selected tab
  const selectedTab = tabs.find((tab) => !!router.pathname.match(tab.selectRegex));
  const classes = useStyles();

  const makeMenuItem = (text: string | JSX.Element, props: any) => {
    return (
      <MenuItem component={props.as ? NakedLink : "a"} {...props}>
        {text}
      </MenuItem>
    );
  };

  const makeButton = (text: string, props: any) => {
    return (
      <Button component={NakedLink} {...props}>
        {text}
      </Button>
    );
  };

  return (
    <AppBar position="sticky" title={"Hello"}>
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
        <Link
          className={classes.logo}
          component={NakedLink}
          href="/"
          variant="h6"
          color="inherit"
          underline="none"
          noWrap
        >
          <BeneathLogo />
        </Link>
        {me && (
          <>
            <Button component={NakedLink} href="/-/sql">
              SQL
            </Button>
            <Button component={NakedLink} href="/-/create/project">
              Create project
            </Button>
            <Button component={NakedLink} href="/-/create/stream">
              Create stream
            </Button>
          </>
        )}
        <div className={classes.grow} />
        <Tabs
          className={classes.tabs}
          indicatorColor="primary"
          textColor="primary"
          value={selectedTab ? selectedTab.href : false}
        >
          {tabs.map(({ href, label, external }) => (
            <Tab key={href} label={label} value={href} component={external ? "a" : NakedLink} href={href} />
          ))}
        </Tabs>
        {/* Login-specific stuff */}
        <>
          {!me &&
            makeButton("Login", {
              color: "inherit",
              size: "small",
              href: "/-/auth",
              as: "/-/auth",
            })}
          {me && (
            <>
              <IconButton edge="end" aria-haspopup="true" onClick={openMenu} color="inherit">
                <Person />
              </IconButton>
              <Menu
                autoFocus={false}
                anchorEl={menuAnchorEl}
                open={isMenuOpen}
                onClose={closeMenu}
                PaperProps={{ className: classes.menuPaper }}
                anchorOrigin={{ vertical: "top", horizontal: "right" }}
                transformOrigin={{ vertical: "top", horizontal: "right" }}
              >
                <MenuItem disabled className={classes.menuItemHeader}>
                  <div>
                    <Typography variant="h4">{me.displayName}</Typography>
                    <Typography variant="subtitle1" gutterBottom>
                      @{me.name}
                    </Typography>
                  </div>
                </MenuItem>
                {me.readQuota &&
                  me.personalUser?.billingOrganizationID === me.organizationID &&
                  makeMenuItem(
                    <UsageIndicator standalone={false} kind="read" usage={me.readUsage} quota={me.readQuota} />,
                    {
                      onClick: closeMenu,
                      as: `/${me.name}/-/monitoring`,
                      href: `/organization?organization_name=${me.name}&tab=monitoring`,
                      className: classes.menuItemUsage,
                    }
                  )}
                {makeMenuItem("Profile", {
                  onClick: closeMenu,
                  as: `/${me.name}`,
                  href: `/organization?organization_name=${me.name}`,
                })}
                {makeMenuItem("Secrets", {
                  onClick: closeMenu,
                  as: `/${me.name}/-/secrets`,
                  href: `/organization?organization_name=${me.name}&tab=secrets`,
                })}
                {me.personalUser &&
                  me.personalUser.billingOrganizationID !== me.organizationID &&
                  makeMenuItem(
                    me.personalUser.billingOrganization.displayName || me.personalUser.billingOrganization.name,
                    {
                      onClick: closeMenu,
                      as: `/${me.personalUser.billingOrganization.name}`,
                      href: `/organization?organization_name=${me.personalUser.billingOrganization.name}`,
                    }
                  )}
                {makeMenuItem("Logout", {
                  href: `/-/redirects/auth/logout`,
                })}
              </Menu>
            </>
          )}
        </>
      </Toolbar>
    </AppBar>
  );
};

export default Header;
