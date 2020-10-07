import { NextRouter, withRouter } from "next/router";
import React, { FC } from "react";

import CircularProgress from "@material-ui/core/CircularProgress";
import Divider from "@material-ui/core/Divider";
import { makeStyles, Theme } from "@material-ui/core/styles";
import Tab from "@material-ui/core/Tab";
import Tabs from "@material-ui/core/Tabs";

import { NakedLink } from "./Link";
import VSpace from "./VSpace";

const useStyles = makeStyles((theme: Theme) => ({
  label: {
    flex: "1",
  },
  labelContainer: {
    display: "inline-flex",
    width: "100%",
  },
  labelProgress: {
    width: "16px",
    marginLeft: "5px",
    marginTop: "3px",
  },
  tabsRoot: {
    [theme.breakpoints.up("lg")]: {
      justifyContent: "center",
    },
  },
  tabsScroller: {
    [theme.breakpoints.up("lg")]: {
      flex: "none",
      justifyContent: "center",
    },
  },
}));

export interface SubrouteTabsProps {
  router: NextRouter;
  tabs: SubrouteTab[];
  defaultValue: string;
}

export interface SubrouteTab {
  value: string;
  label: string;
  render: () => React.ReactNode;
}

const buildHref = (path: string, query: NodeJS.Dict<string>) => {
  const queryString = Object.keys(query)
    .map((key) => encodeURIComponent(key) + "=" + encodeURIComponent(query[key] || ""))
    .join("&");
  return `${path}?${queryString}`;
};

const SubrouteTabs: FC<SubrouteTabsProps> = ({ router, tabs, defaultValue }) => {
  const selectedValue = router.query.tab || defaultValue || tabs[0].value;
  const asPathBase = router.query.tab ? router.asPath.substring(0, router.asPath.lastIndexOf("/-/")) : router.asPath;
  const selectedTab = tabs.find((tab) => tab.value === selectedValue);
  const [loading, setLoading] = React.useState(false);
  const classes = useStyles();
  return (
    <React.Fragment>
      <Tabs
        indicatorColor="primary"
        textColor="primary"
        scrollButtons="auto"
        value={selectedValue}
        variant="scrollable"
        classes={{ root: classes.tabsRoot, scroller: classes.tabsScroller }}
      >
        {tabs.map((tab) => (
          <Tab
            key={tab.value}
            value={tab.value}
            label={
              <div className={classes.labelContainer}>
                <span className={classes.label}>{tab.label}</span>
                {tab.value === selectedValue && loading && (
                  <CircularProgress className={classes.labelProgress} size={16} disableShrink />
                )}
              </div>
            }
            component={NakedLink}
            shallow
            replace
            as={`${asPathBase}/-/${tab.value}`}
            href={buildHref(router.pathname, { ...router.query, tab: tab.value })}
          />
        ))}
      </Tabs>
      <Divider />
      <VSpace units={4} />
      {selectedTab && selectedTab.render()}
    </React.Fragment>
  );
};

export default withRouter(SubrouteTabs);
