import React, { FC } from "react";

import useMe from "../../hooks/useMe";
import BeneathLogo from "./BeneathLogo";
import { NakedLink } from "../Link";

import {
  AppBar,
  Button,
  Link,
  makeStyles,
  Toolbar,
  useMediaQuery,
  useTheme,
} from "@material-ui/core";

import { Menu as MenuIcon, Add } from "@material-ui/icons";
import PathBreadcrumbs from "./PathBreadcrumbs";
import ProfileButton from "./ProfileButton";
import SplitButton from "components/SplitButton";
import DropdownButton from "components/DropdownButton";
import clsx from "clsx";

const useStyles = makeStyles((_) => ({
  grow: {
    flexGrow: 1,
  },
  logo: {
    height: "28px",
  },
  rightItem: {
    marginLeft: "0.5rem",
  },
  rightButton: {
    height: "32px",
  },
}));

const Header: FC = () => {
  const me = useMe();
  const classes = useStyles();

  const theme = useTheme();
  const isSm = useMediaQuery(theme.breakpoints.up("md"));

  const createActions = [
    { label: "Create project", href: "/-/create/project" },
    { label: "Create stream", href: "/-/create/stream" },
    { label: "Create service", href: "/-/create/service" },
  ] as { label: string; href: string; as?: string }[];
  if (me) {
    createActions.push({
      label: "Create personal secret",
      href: `/organization?organization_name=${me.name}&tab=secrets`,
      as: `/${me.name}/-/secrets`,
    });
  }

  const linkActions = [{ label: "Docs", href: "https://about.beneath.dev/docs/" }];
  if (me) {
    linkActions.unshift({ label: "SQL", href: "/-/sql" });
  }

  return (
    <AppBar position="sticky" title="Beneath">
      <Toolbar variant="dense">
        {/* Logo */}
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
        {/* Path */}
        <PathBreadcrumbs />
        {/* Change to right-hand side */}
        <div className={classes.grow} />
        {/* Create stream/project/etc. button */}
        {me &&
          (isSm ? (
            <SplitButton
              margin="dense"
              className={clsx(classes.rightItem, classes.rightButton)}
              color="secondary"
              variant="contained"
              mainActionIdx={1}
              actions={createActions}
            />
          ) : (
            <DropdownButton
              className={clsx(classes.rightItem, classes.rightButton)}
              color="secondary"
              variant="contained"
              margin="dense"
              actions={createActions}
            >
              <Add />
            </DropdownButton>
          ))}
        {/* Links (desktop) */}
        {isSm &&
          linkActions.map((action, idx) => (
            <Button
              key={idx}
              className={clsx(classes.rightItem, classes.rightButton)}
              component={NakedLink}
              href={action.href}
            >
              {action.label}
            </Button>
          ))}

        {/* Login button */}
        {!me && (
          <Button
            className={clsx(classes.rightItem, classes.rightButton)}
            component={NakedLink}
            variant="contained"
            href="/-/auth"
          >
            Login
          </Button>
        )}
        {/* Signup button */}
        {!me && (
          <Button
            className={clsx(classes.rightItem, classes.rightButton)}
            component={NakedLink}
            variant="contained"
            href="/-/auth"
            color="primary"
          >
            Sign up
          </Button>
        )}
        {me && <ProfileButton className={classes.rightItem} me={me} />}
        {/* Links (mobile) */}
        {!isSm && (
          <DropdownButton
            className={clsx(classes.rightItem, classes.rightButton)}
            variant="text"
            margin="dense"
            actions={linkActions}
          >
            <MenuIcon />
          </DropdownButton>
        )}
      </Toolbar>
    </AppBar>
  );
};

export default Header;
