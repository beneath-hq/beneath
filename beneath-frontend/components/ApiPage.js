import React, { Component } from "react";
import { withRouter } from "next/router";
import { AuthRequired } from "../hocs/auth";
import PageTitle from "./PageTitle";
import App from "./App";
import { ApiSidebar } from "./Sidebar";
import { BurgerMenu, SideDrawer } from "./BurgerMenu";
import { devices } from "../lib/theme";

class ApiPageHeader extends Component {
  state = {
    sideDrawerOpen: false,
  };

  drawerToggleClickHandler = () => {
    this.setState(prevState => {
      return { sideDrawerOpen: !prevState.sideDrawerOpen };
    });
  };

  render() {
    let sideDrawer;

    if (this.state.sideDrawerOpen) {
      sideDrawer = <SideDrawer show={this.state.sideDrawerOpen} />;
    }

    return (
      <div className="page-header">
        <div className="burger-menu-container">
          <BurgerMenu clickHandler={this.drawerToggleClickHandler} />
        </div>
        <div className="breadcrumbs-container">
          {this.props.router.pathname}
        </div>
        {sideDrawer}
        <style jsx>{`
        .page-header {
          display: flex;
          align-items: center;
          height: 35px;
          padding-left: 10px;
          border-bottom: 1px solid #405070;
        }

        .breadcrumbs-container {
          padding-left: 10px;
          display: flex;
        }

        @media ${devices.smallerThanTablet} {
          div.burger-menu-container {
            display: flex;
          }
        }

        @media ${devices.tabletOrLarger} {
          div.burger-menu-container {
            display: none;
          }
        }
        `}</style>
      </div>
    );
  }
}

const ApiPageHeaderWithRouter = withRouter(ApiPageHeader);

const ApiPage = props => {
  return (
    <div>
      <App>
        <PageTitle subtitle="API" />
        <div className="api-container">
          <div className="api-sidebar">
            <ApiSidebar />
          </div>
          <div className="main-content">
            <ApiPageHeaderWithRouter />
            <div className="api-page">{props.children}</div>
          </div>
        </div>
      </App>
      <style jsx>{`
        .api-container {
          display: flex;
          flex-grow: 1;
        }

        .main-content {
          display: flex;
          flex-direction: column;
          flex-grow: 1;
        }

        div.api-page {
          display: flex;
          flex-grow: 1;
        }

        @media ${devices.smallerThanTablet} {
          div.api-sidebar {
            display: none;
          }
        }

        @media ${devices.tabletOrLarger} {
          div.api-sidebar {
            display: flex;
          }
        }
      `}</style>
    </div>
  );
};

export default ApiPage;
