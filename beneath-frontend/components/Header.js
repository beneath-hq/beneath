import AppBar from "@material-ui/core/AppBar";
import Button from "@material-ui/core/Button";
import IconButton from "@material-ui/core/IconButton";
import Link from "@material-ui/core/Link";
import Menu from "@material-ui/core/Menu";
import MenuIcon from "@material-ui/icons/Menu";
import MenuItem from "@material-ui/core/MenuItem";
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
  { label: "Explore", href: "/explore", selectRegex: "/(explore|project).*" },
  { label: "About", href: "/about", selectRegex: "/about.*" },
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
  // prepare profile menu
  const [menuAnchorEl, setMenuAnchorEl] = React.useState(null);
  const isMenuOpen = !!menuAnchorEl;
  const openMenu = (event) => setMenuAnchorEl(event.currentTarget);
  const closeMenu = () => setMenuAnchorEl(null);

  // compute selected tab
  const selectedTab = tabs.find((tab) => !!router.pathname.match(tab.selectRegex));
  const classes = useStyles();
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
        <Tabs value={selectedTab ? selectedTab.href : false}>
          {tabs.map(({ href, label }) => (
            <Tab key={href} label={label} value={href} component={NextMuiLink} href={href} />
          ))}
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
                  <React.Fragment>
                    <IconButton edge="end" aria-haspopup="true" onClick={openMenu} color="inherit">
                      <Person />
                    </IconButton>
                    <Menu anchorEl={menuAnchorEl} open={isMenuOpen} onClose={closeMenu}
                      anchorOrigin={{ vertical: "top", horizontal: "right" }}
                      transformOrigin={{ vertical: "top", horizontal: "right" }}
                    >
                      <MenuItem onClick={closeMenu} component={NextMuiLink} href="/users/me">Profile</MenuItem>
                      <MenuItem component={NextMuiLink} href="/auth/logout">Logout</MenuItem>
                    </Menu>
                  </React.Fragment>
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
