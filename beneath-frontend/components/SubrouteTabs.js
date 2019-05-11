import PropTypes from "prop-types";
import React, { Component } from "react";
import { withRouter } from "next/router";
import Tabs from "@material-ui/core/Tabs";
import Tab from "@material-ui/core/Tab";
import Divider from "@material-ui/core/Divider";

import NextMuiLink from "./NextMuiLink";

const SubrouteTabs = ({ router, tabs, defaultValue }) => {
  let selectedValue = router.query.tab || defaultValue || tabs[0].value;
  let asPathBase = router.query.tab ? router.asPath.substring(0, router.asPath.lastIndexOf("/")) : router.asPath;
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
            label={tab.label}
            component={NextMuiLink}
            shallow
            replace
            as={`${asPathBase}/${tab.value}`}
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
      {tabs.find((tab) => tab.value === selectedValue).render()}
    </React.Fragment>
  );
};

SubrouteTabs.propTypes = {
  defaultValue: PropTypes.string,
  tabs: PropTypes.arrayOf(
    PropTypes.shape({
      value: PropTypes.string.isRequired,
      label: PropTypes.string.isRequired,
      render: PropTypes.func.isRequired,
    }).isRequired
  ).isRequired,
};

export default withRouter(SubrouteTabs);
