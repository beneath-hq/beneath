import { Container, Grid, makeStyles, Typography } from "@material-ui/core";
import Head from "next/head";
import React from "react";

import Drawer from "./Drawer";
import Header from "./header/Header";
import { Link } from "./Link";

interface IStylesProps {
  contentMarginTop?: null | "dense" | "normal" | "hero";
}

const useStyles = makeStyles((theme) => ({
  container: {
    minHeight: "100vh",
    position: "relative",
  },
  footer: {
    position: "absolute",
    bottom: 0,
    width: "100%",
    padding: theme.spacing(3),
    textAlign: "center",
  },
  sidebarSubheaderAndContent: {
    display: "flex",
  },
  subheaderAndContent: {
    overflow: "auto",
    flexGrow: 1,
  },
  content: ({ contentMarginTop }: IStylesProps) => ({
    marginBottom: theme.spacing(12),
    marginTop:
      contentMarginTop === "normal"
        ? theme.spacing(6)
        : contentMarginTop === "dense"
        ? theme.spacing(3)
        : contentMarginTop === "hero"
        ? theme.spacing(10)
        : theme.spacing(0),
  }),
}));

interface IProps {
  title?: string;
  sidebar?: JSX.Element;
  maxWidth?: false | "xs" | "sm" | "md" | "lg" | "xl";
  contentMarginTop?: null | "dense" | "normal" | "hero";
}

const Page: React.FC<IProps> = (props) => {
  const [mobileDrawerOpen, setMobileDrawerOpen] = React.useState(false);
  const toggleMobileDrawer = () => {
    setMobileDrawerOpen(!mobileDrawerOpen);
  };

  const title = props.title ? props.title + " | Beneath" : "Beneath";
  const classes = useStyles({ contentMarginTop: props.contentMarginTop });
  return (
    <div className={classes.container}>
      <Head>
        <title>{title}</title>
        <meta name="twitter:card" content="summary" />
        <meta name="twitter:site" content="@BeneathHQ" />
        <meta property="og:image" content="https://about.beneath.dev/media/logo/banner-square.png" />
        <meta property="og:title" content={title} />
      </Head>
      <Header />
      <div className={classes.sidebarSubheaderAndContent}>
        {props.sidebar && (
          <Drawer mobileOpen={mobileDrawerOpen} toggleMobileOpen={toggleMobileDrawer}>
            {props.sidebar}
          </Drawer>
        )}
        <div className={classes.subheaderAndContent}>
          <Container maxWidth={props.maxWidth === undefined ? "lg" : props.maxWidth}>
            <main className={classes.content}>{props.children}</main>
          </Container>
        </div>
      </div>
      <div className={classes.footer}>
        <Grid container spacing={4} justify="center">
          <Grid item>&copy; 2021 Beneath</Grid>
          <Grid item>
            <Link href="https://about.beneath.dev/policies/terms/" target="_blank">
              Terms of Service
            </Link>
          </Grid>
          <Grid item>
            <Link href="https://about.beneath.dev/policies/privacy/" target="_blank">
              Privacy Policy
            </Link>
          </Grid>
          <Grid item>
            <Link href="https://about.beneath.dev/contact/" target="_blank">
              Contact us
            </Link>
          </Grid>
        </Grid>
      </div>
    </div>
  );
};

export default Page;
