import { useRouter } from "next/router";
import React, { FC } from "react";

import useMe from "../hooks/useMe";
import UsageIndicator from "./metrics/user/UsageIndicator";
import { NakedLink } from "./Link";

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
          {/* Pasted and adapted from "icon-text-white.svg" */}
          <svg viewBox="0 0 475 95" height="100%">
            <title>Beneath Logo</title>
            <path
              d="M 124.784 73.959 L 124.784 24.399 L 146.724 24.399 C 150.837 24.399 154.067 25.559 156.414 27.879 C 158.754 30.199 159.924 33.299 159.924 37.179 C 159.924 39.026 159.664 40.599 159.144 41.899 C 158.624 43.206 157.961 44.272 157.154 45.099 C 156.354 45.926 155.431 46.539 154.384 46.939 C 153.344 47.346 152.327 47.596 151.334 47.689 L 151.334 48.119 C 152.327 48.166 153.427 48.402 154.634 48.829 C 155.841 49.249 156.964 49.922 158.004 50.849 C 159.051 51.769 159.927 52.952 160.634 54.399 C 161.347 55.839 161.704 57.602 161.704 59.689 C 161.704 61.676 161.384 63.546 160.744 65.299 C 160.104 67.046 159.217 68.559 158.084 69.839 C 156.944 71.119 155.594 72.126 154.034 72.859 C 152.474 73.592 150.771 73.959 148.924 73.959 L 124.784 73.959 Z M 132.804 52.019 L 132.804 67.069 L 146.574 67.069 C 148.661 67.069 150.294 66.526 151.474 65.439 C 152.661 64.352 153.254 62.789 153.254 60.749 L 153.254 58.339 C 153.254 56.306 152.661 54.742 151.474 53.649 C 150.294 52.562 148.661 52.019 146.574 52.019 L 132.804 52.019 Z M 132.804 31.289 L 132.804 45.349 L 145.224 45.349 C 147.217 45.349 148.757 44.839 149.844 43.819 C 150.931 42.799 151.474 41.342 151.474 39.449 L 151.474 37.179 C 151.474 35.286 150.931 33.832 149.844 32.819 C 148.757 31.799 147.217 31.289 145.224 31.289 L 132.804 31.289 ZM 209.303 73.959 L 177.573 73.959 L 177.573 24.399 L 209.303 24.399 L 209.303 31.499 L 185.593 31.499 L 185.593 45.279 L 207.103 45.279 L 207.103 52.379 L 185.593 52.379 L 185.593 66.859 L 209.303 66.859 L 209.303 73.959 ZM 255.206 73.959 L 239.016 46.549 L 233.556 36.049 L 233.336 36.049 L 233.336 73.959 L 225.676 73.959 L 225.676 24.399 L 234.616 24.399 L 250.806 51.809 L 256.276 62.319 L 256.486 62.319 L 256.486 24.399 L 264.156 24.399 L 264.156 73.959 L 255.206 73.959 ZM 314.1 73.959 L 282.37 73.959 L 282.37 24.399 L 314.1 24.399 L 314.1 31.499 L 290.39 31.499 L 290.39 45.279 L 311.9 45.279 L 311.9 52.379 L 290.39 52.379 L 290.39 66.859 L 314.1 66.859 L 314.1 73.959 ZM 369.803 73.959 L 361.423 73.959 L 356.953 60.539 L 338.423 60.539 L 334.093 73.959 L 325.923 73.959 L 342.823 24.399 L 352.903 24.399 L 369.803 73.959 Z M 354.963 53.649 L 347.863 31.789 L 347.513 31.789 L 340.343 53.649 L 354.963 53.649 ZM 412.01 31.499 L 397.25 31.499 L 397.25 73.959 L 389.22 73.959 L 389.22 31.499 L 374.45 31.499 L 374.45 24.399 L 412.01 24.399 L 412.01 31.499 ZM 456.287 73.959 L 456.287 52.379 L 433.847 52.379 L 433.847 73.959 L 425.827 73.959 L 425.827 24.399 L 433.847 24.399 L 433.847 45.279 L 456.287 45.279 L 456.287 24.399 L 464.307 24.399 L 464.307 73.959 L 456.287 73.959 Z"
              fill="rgb(255, 255, 255)"
              transform="matrix(1, 0, 0, 1, 0, 0)"
            />
            <g transform="matrix(-1, 0, 0, -1, 208.382507, 172.398163)">
              <path
                d="M 160.538 91.938 C 168.372 91.938 175.718 94.033 182.044 97.694 C 182.051 105.003 180.193 112.414 176.276 119.198 C 172.358 125.983 166.87 131.298 160.538 134.945 C 154.203 131.298 148.715 125.983 144.799 119.198 C 140.881 112.414 139.022 105.003 139.03 97.694 C 145.357 94.033 152.702 91.938 160.538 91.938 Z"
                fillOpacity={0.71}
                fill="none"
                strokeWidth="7px"
                paintOrder="fill"
                stroke="rgb(255, 255, 255)"
                mask="url(#mask-2)"
                transform="matrix(-0.999848, 0.017452, -0.017452, -0.999848, 323.029321, 224.064043)"
              />
              <path
                d="M 161.138 129.185 C 174.919 137.141 182.629 151.575 182.644 166.42 C 176.318 162.759 168.973 160.664 161.138 160.664 C 153.302 160.664 145.958 162.759 139.63 166.42 C 139.645 151.575 147.357 137.141 161.138 129.185 Z"
                fillOpacity={0.71}
                fill="none"
                strokeWidth="7px"
                paintOrder="fill"
                stroke="rgb(255, 255, 255)"
                mask="url(#mask-3)"
                transform="matrix(-0.999848, 0.017452, -0.017452, -0.999848, 324.82891, 292.770323)"
              />
              <path
                d="M 160.44 123.797 C 146.658 131.754 130.303 131.214 117.44 123.805 C 123.774 120.157 129.261 114.842 133.178 108.057 C 137.095 101.272 138.953 93.863 138.946 86.554 C 151.795 93.99 160.439 107.884 160.44 123.797 Z"
                fillOpacity={0.71}
                fill="none"
                strokeWidth="7px"
                paintOrder="fill"
                stroke="rgb(255, 255, 255)"
                mask="url(#mask-4)"
                transform="matrix(-0.999848, 0.017452, -0.017452, -0.999848, 279.744699, 213.677735)"
              />
              <path
                d="M 160.433 123.046 C 160.434 107.134 169.078 93.24 181.927 85.804 C 181.918 93.112 183.777 100.522 187.695 107.307 C 191.611 114.092 197.098 119.407 203.433 123.055 C 190.569 130.464 174.214 131.004 160.433 123.046 Z"
                fillOpacity={0.71}
                fill="none"
                strokeWidth="7px"
                paintOrder="fill"
                stroke="rgb(255, 255, 255)"
                mask="url(#mask-5)"
                transform="matrix(-0.999848, 0.017452, -0.017452, -0.999848, 365.711053, 211.427413)"
              />
            </g>
          </svg>
        </Link>
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
