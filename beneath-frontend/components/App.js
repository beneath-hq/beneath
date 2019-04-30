import AppBar from "@material-ui/core/AppBar";
import Button from "@material-ui/core/Button";
import IconButton from "@material-ui/core/IconButton";
import Person from "@material-ui/icons/Person";
import Toolbar from "@material-ui/core/Toolbar";
import Typography from "@material-ui/core/Typography";
import Tabs from "@material-ui/core/Tabs";
import Tab from "@material-ui/core/Tab";
import { makeStyles } from "@material-ui/core/styles";
import { withRouter } from "next/router";

import { AuthConsumer } from "../hocs/auth";
import NextMuiLink from "./NextMuiLink";

const useStyles = makeStyles((theme) => ({
  grow: {
    flexGrow: 1,
  },
}));

const Header = withRouter((({ router }) => {
  const classes = useStyles();
  return (
    <AppBar position="relative" colorPrimary="rgba(16, 24, 46, 1)">
      <Toolbar variant="dense">
        <Typography variant="h6" color="inherit" noWrap>
          Material-UI
        </Typography>
        <div className={classes.grow} />
        <Tabs value={router.pathname}>
          <Tab label="About" value="/about" component={NextMuiLink} href="/about" />
          <Tab label="Explore" value="/explore" component={NextMuiLink} href="/explore" />
        </Tabs>
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
}));

export default ({ children }) => (
  <div className="main">
    <Header />
    {children}
  </div>
);
