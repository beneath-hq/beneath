import AppBar from "@material-ui/core/AppBar";
import Button from "@material-ui/core/Button";
import IconButton from "@material-ui/core/IconButton";
import Link from "@material-ui/core/Link";
import MenuIcon from "@material-ui/icons/Menu";
import Person from "@material-ui/icons/Person";
import Toolbar from "@material-ui/core/Toolbar";
import Tabs from "@material-ui/core/Tabs";
import Tab from "@material-ui/core/Tab";
import { makeStyles } from "@material-ui/core/styles";
import { withRouter } from "next/router";
import PropTypes from "prop-types";

import NextMuiLink from "./NextMuiLink";
import { AuthConsumer } from "../hocs/auth";

const tabs = [
  { label: "Explore", value: "/explore", href: "/explore" },
  { label: "About", value: "/about", href: "/about" },
];

const useStyles = makeStyles((theme) => ({
  grow: {
    flexGrow: 1,
  },
  drawerButton: {
    marginRight: theme.spacing(2),
    marginLeft: 0,
    [theme.breakpoints.up('sm')]: {
      display: 'none',
    },
  },
}));

const Header = (({ router, toggleMobileDrawer }) => {
  const classes = useStyles();
  const selectedTab = tabs.find((tab) => router.pathname.startsWith(tab.value));
  return (
    <AppBar position="relative" className={classes.appBar}>
      <Toolbar variant="dense">
        {toggleMobileDrawer && (
          <IconButton color="inherit" edge="start" aria-label="Open drawer"
            className={classes.drawerButton} onClick={toggleMobileDrawer}
          >
            <MenuIcon />
          </IconButton>
        ) }
        <Link component={NextMuiLink} href="/" variant="h6" color="inherit" underline="none" noWrap>
          BENEATH
        </Link>
        <div className={classes.grow} />
        <Tabs value={selectedTab ? selectedTab.value : false}>
          {tabs.map((tab) => <Tab key={tab.value} component={NextMuiLink} {...tab} />)}
        </Tabs>
        {/* Login-specific stuff */}
        <AuthConsumer>
          {({ user }) => {
            return (
              <React.Fragment>
                {!user && (
                  <Button color="inherit" component={NextMuiLink} size="small" href="/auth">Login</Button>
                )}
                {user && (
                  <IconButton color="inherit" component={NextMuiLink} href="/profile">
                    <Person />
                  </IconButton>
                )}
              </React.Fragment>
            );
          }}
        </AuthConsumer>
      </Toolbar>
    </AppBar>
  );
});

Header.propTypes = {
  toggleMobileDrawer: PropTypes.func,
};

export default withRouter(Header);
