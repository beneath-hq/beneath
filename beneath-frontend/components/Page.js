import PropTypes from "prop-types";
import React from "react";
import Container from "@material-ui/core/Container";
import { makeStyles } from "@material-ui/core/styles";

import Drawer from "./Drawer";
import Header from "./Header";
import PageTitle from "./PageTitle";
import Subheader from "./Subheader";

const useStyles = makeStyles((theme) => ({
  sidebarSubheaderAndContent: {
    display: "flex"
  },
  subheaderAndContent: {
    flexGrow: 1,
    padding: theme.spacing(3),
  },
  content: ({ contentMarginTop }) => ({
    marginTop: (
      contentMarginTop === "normal"
        ? theme.spacing(6)
        : contentMarginTop === "dense"
        ? theme.spacing(3)
        : contentMarginTop === "hero"
        ? theme.spacing(10)
        : theme.spacing(0)
    ),
  }),
}));

const Page = (props) => {
  const [mobileDrawerOpen, setMobileDrawerOpen] = React.useState(false);
  const toggleMobileDrawer = () => {
    setMobileDrawerOpen(!mobileDrawerOpen);
  };

  const classes = useStyles({ contentMarginTop: props.contentMarginTop });
  return (
    <div>
      <PageTitle title={props.title} />
      <Header toggleMobileDrawer={props.sidebar && toggleMobileDrawer} />
      <div className={classes.sidebarSubheaderAndContent}>
        { props.sidebar && (
          <Drawer mobileOpen={mobileDrawerOpen} toggleMobileOpen={toggleMobileDrawer}>
            {props.sidebar}
          </Drawer>
        )}
        <div className={classes.subheaderAndContent}>
          <Container maxWidth={props.maxWidth || "lg"}>
            { props.sidebar && <Subheader /> }
            <main className={classes.content}>
              {props.children}
            </main>
          </Container>
        </div>
      </div>
    </div>
  );
};

Page.propTypes = {
  title: PropTypes.string,
  sidebar: PropTypes.object,
  maxWidth: PropTypes.oneOf([false, "xs", "sm", "md", "lg", "xl"]),
  contentMarginTop: PropTypes.oneOf([null, "dense", "normal"]),
};

export default Page;
