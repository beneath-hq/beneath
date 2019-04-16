import App from "../components/App";
import PageTitle from "../components/PageTitle";
import PropTypes from "prop-types";
import { devices } from "../lib/theme";
import { withRouter } from "next/router";

import React, { Component } from "react";

export default class Page extends Component {
  static propTypes = {
    title: PropTypes.string,
    sidebar: PropTypes.object,
  };

  constructor(props) {
    super(props);
  }

  render() {
    let content = null;
    if (this.props.sidebar) {
      content = SidebarContent(this.props);
    } else {
      content = PlainContent(this.props);
    }

    return (
      <App>
        <PageTitle subtitle={this.props.title} />
        { content }
      </App>
    );
  }
}

const PlainContent = (props) => {
  return props.children;
};

const SidebarContent = (props) => {
  return (
    <div className="container">
      <div className="left-container">
        { props.sidebar }
      </div>
      <div className="right-container">
        <SidebarContentHeader sidebar={props.sidebar} />
        <div className="content-container">{props.children}</div>
      </div>
      <style jsx>{`
        .container {
          display: flex;
          flex-grow: 1;
        }

        .right-container {
          display: flex;
          flex-direction: column;
          flex-grow: 1;
        }

        .content-container {
          display: flex;
          flex-grow: 1;
        }

        @media ${devices.smallerThanTablet} {
          .left-container {
            display: none;
          }
        }

        @media ${devices.tabletOrLarger} {
          .left-container {
            display: flex;
          }
        }
      `}</style>
    </div>
  );
};

const SidebarContentHeader = withRouter(class extends Component {
  constructor(props) {
    super(props);
    this.state = {
      sideDrawerOpen: false,
    };
  }

  toggleDrawer() {
    this.setState((prevState) => {
      return {
        sideDrawerOpen: !prevState.sideDrawerOpen
      };
    });
  };

  render() {
    let sideDrawer = null;
    if (this.state.sideDrawerOpen) {
      sideDrawer = (
        <SideDrawer>
          {this.props.sidebar}
        </SideDrawer>
      );
    }

    return (
      <div className="page-header">
        <div className="burger-menu-container">
          <BurgerIcon onClick={this.toggleDrawer.bind(this)} />
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
});

const BurgerIcon = (props) => (
  <div className="container" onClick={(event) => {
    // call onClick
    props.onClick();
    // animate the burger icon
    event.currentTarget.classList.toggle("change");
  }}>
    <div className="bar1" />
    <div className="bar2" />
    <div className="bar3" />
    <style jsx>{`
      .container {
        cursor: pointer;
      }
      .bar1, .bar2, .bar3 {
        width: 35px;
        height: 5px;
        background-color: #50c0ff;
        margin: 6px 0;
        transition: 0.4s;
      }
      .change .bar1 {
        transform: rotate(-45deg) translate(-9px, 6px);
      }
      .change .bar2 {
        opacity: 0;
      }
      .change .bar3 {
        transform: rotate(45deg) translate(-8px, -8px);
      }
    `}</style>
  </div>
);

const SideDrawer = (props) => {
  return (
    <div>
      <div className={`side-drawer`}>
        {props.children}
      </div>
      <style jsx>{`
        .side-drawer {
          position: fixed;
          top: 97px;
          left: 0;
          width: 100%;
          height: 90%;
          z-index: 2;

          background: #101830;
          box-shadow: 1px 0px 7px rgba(0, 0, 0, 0.5);
          
          /* // Goodies to explore some other time
            width: 70%;
            max-width: 400px;
            animation: 0.3s ease-out 0s 1 slideInFromLeft;
          */
        }
      `}</style>
    </div>
  );
};
