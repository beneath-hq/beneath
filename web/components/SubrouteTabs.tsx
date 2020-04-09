import { NextRouter, withRouter } from "next/router";
import PropTypes from "prop-types";
import React, { Component, FC } from "react";

import CircularProgress from "@material-ui/core/CircularProgress";
import Divider from "@material-ui/core/Divider";
import { makeStyles, Theme } from "@material-ui/core/styles";
import Tab from "@material-ui/core/Tab";
import Tabs from "@material-ui/core/Tabs";

import NextMuiLink from "./NextMuiLink";
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
}));

export interface SubrouteTabsProps {
  router: NextRouter;
  tabs: SubrouteTab[];
  defaultValue: string;
}

export interface SubrouteTab {
  value: string;
  label: string;
  render: (props: SubrouteTabProps) => React.ReactNode;
}

export interface SubrouteTabProps {
  setLoading: (loading: boolean) => void;
}

const SubrouteTabs: FC<SubrouteTabsProps> = ({ router, tabs, defaultValue }) => {
  const selectedValue = router.query.tab || defaultValue || tabs[0].value;
  const asPathBase = router.query.tab ? router.asPath.substr(0, router.asPath.lastIndexOf("/", router.asPath.lastIndexOf("/") - 1)) : router.asPath;
  const selectedTab = tabs.find((tab) => tab.value === selectedValue);
  const [loading, setLoading] = React.useState(false);
  const classes = useStyles();
  return (
    <React.Fragment>
      <Tabs
        indicatorColor="primary"
        textColor="primary"
        variant="scrollable"
        scrollButtons="off"
        value={selectedValue}
      >
        {tabs.map((tab) => (
          <Tab
            key={tab.value}
            value={tab.value}
            label={
              <div className={classes.labelContainer}>
                <span className={classes.label}>{tab.label}</span>
                {tab.value === selectedValue && loading && (
                  <CircularProgress className={classes.labelProgress} size={16} />
                )}
              </div>
            }
            component={NextMuiLink}
            shallow
            replace
            as={`${asPathBase}/-/${tab.value}`}
            href={{
              pathname: router.pathname,
              query: {
                ...router.query,
                tab: tab.value,
              },
            }}
          />
        ))}
      </Tabs>
      <Divider />
      <VSpace units={4} />
      {selectedTab && selectedTab.render({ setLoading })}
    </React.Fragment>
  );
};

export default withRouter(SubrouteTabs);
