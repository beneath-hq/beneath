import Link from "next/link";
import { withRouter } from "next/router";
import { AuthConsumer } from "../hocs/auth";
import { devices } from "../lib/theme";

const Nav = props => {
  return (
    <nav>
      {props.children}
      <style jsx>{`
        nav {
          height: 60px;
          border-bottom: 1px solid #405070;
          display: flex;
          flex-grow: 1;
          align-items: stretch;
          justify-content: space-between;
        }
      `}</style>
    </nav>
  );
};

const NavLeft = props => {
  return (
    <div>
      {props.children}
      <style jsx>{`
        div {
          display: flex;
          align-items: stretch;
          justify-content: flex-start;
        }
      `}</style>
    </div>
  );
};

const NavRight = props => {
  return (
    <div>
      {props.children}
      <style jsx>{`
        div {
          display: flex;
          align-items: stretch;
          justify-content: flex-end;
        }
        @media ${devices.smallerThanTablet} {
          div {
            flex-grow: 1;
          }
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
          align-items: center;
          text-align: center;
        }
        @media ${devices.smallerThanTablet} {
          div {
            flex-grow: 1;
          }
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
            border-left: 1px solid #405070;
            padding: 0px 40px;
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

const Brand = () => {
  return (
    <div>
      <h1 className="text">Beneath</h1>
      <h1 className="icon">&#8962;</h1>
      <style jsx>{`
        div {
          flex-grow: 1;
        }
        @media ${devices.smallerThanTablet} {
          h1.text {
            display: none;
          }
          div {
            padding-left: 1px
          }
        }
        h1.text {
          margin-left: 7px;
          font-weight: 300;
          font-size: 2rem;
          letter-spacing: 0.1rem;
        }
        @media ${devices.tabletOrLarger} {
          h1.icon {
            display: none;
          }
        }
        h1.icon {
          margin: 10px;
          font-weight: 300;
          font-size: 3.3rem;
          letter-spacing: 0.1rem;
        }
      `}</style>
    </div>
  );
};

const Header = ({ router }) => (
  <div className="header">
    <Nav>
      <NavLeft>
        <NavItem>
          <Link href="/">
            <a>
              <Brand />
            </a>
          </Link>
        </NavItem>
      </NavLeft>

      <AuthConsumer>
        {({ user }) => {
          return (
            <NavRight>
              {user && (
                <NavItem>
                  <NavItemLink href="https://lab.beneath.network/" text="Lab" />
                </NavItem>
              )}
              <NavItem>
                <NavItemLink href={user ? `/auth/logout` : "/auth"} text={user ? "Log out" : "Log in"} />
              </NavItem>
            </NavRight>
          );
        }}
      </AuthConsumer>
    </Nav>
  </div>
);

export default withRouter(Header);
