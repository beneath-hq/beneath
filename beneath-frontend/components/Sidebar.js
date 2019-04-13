import Link from "next/link";
import { withRouter } from "next/router";

const NavList = props => {
  return (
    <div className="nav-list">
      {props.children}
      <style jsx>{`
        div.nav-list {
          display: flex;
          flex-direction: column;
          flex-grow: 1;
          border-right: 1px solid #405070;
        }
      `}</style>
    </div>
  );
};

const NavItem = props => {
  return (
    <div>
      {props.children}
      <style jsx>{`
        div {
          display: flex;
          align-items: stretch;
          justify-content: flex-end;
        }
      `}</style>
    </div>
  );
};

const NavItemLink = props => {
  const { text, pathname, ...linkProps } = props;

  return (
    <Link {...linkProps}>
      <a className={pathname == linkProps.href ? "is-active" : ""}>
        {text}
        <style jsx>{`
          a {
            height: 100%;
            width: 100%;
            display: flex;
            align-items: center;
            justify-content: center;
            border-bottom: 1px solid #405070;
          }
          a:hover,
          a:active,
          .is-active {
            color: #50c0ff;
            background: hsla(0, 0%, 100%, 0.1);
          }
        `}</style>
      </a>
    </Link>
  );
};

export const SidebarLink = ({ href, text }) => (
  <NavItem>
    <NavItemLink href={href} text={text} />
  </NavItem>
);

export const Sidebar = withRouter(({ router: { pathname }, children }) => (
  <NavList>{children}</NavList>
));

export const ApiSidebar = props => (
  <Sidebar>
    <SidebarLink href="/api" text="Dashboard" />
    <SidebarLink href="/api/ping" text="Ping" />
    <SidebarLink href="/api/me" text="Me" />
  </Sidebar>
);
