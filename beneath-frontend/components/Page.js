import PropTypes from "prop-types";
import React from "react";
import Head from "next/head";
import Container from "@material-ui/core/Container";
import { makeStyles } from "@material-ui/core/styles";

import Drawer from "./Drawer";
import Header from "./Header";
import Subheader from "./Subheader";

const useStyles = makeStyles((theme) => ({
  root: {
    display: "flex"
  },
  content: {
    flexGrow: 1,
    padding: theme.spacing(3),
  },
}));

const Page = (props) => {
  const [mobileDrawerOpen, setMobileDrawerOpen] = React.useState(false);
  const toggleMobileDrawer = () => {
    setMobileDrawerOpen(!mobileDrawerOpen);
  };

  const classes = useStyles();
  return (
    <div>
      <Head>
        <title>
          {props.title ? props.title + " | " : ""} Beneath – Data Science for the Decentralised Economy
        </title>
      </Head>
      <Header toggleMobileDrawer={props.sidebar && toggleMobileDrawer} />
      <div className={classes.root}>
        { props.sidebar && (
          <Drawer mobileOpen={mobileDrawerOpen} toggleMobileOpen={toggleMobileDrawer}>
            {props.sidebar}
          </Drawer>
        )}
        <div className={classes.content}>
          <main>
            <Container maxWidth="lg">
              { props.sidebar && <Subheader /> }
              {props.children}
            </Container>
          </main>
        </div>
      </div>
    </div>
  );
};

Page.propTypes = {
  title: PropTypes.string,
  sidebar: PropTypes.object,
};

export default Page;
