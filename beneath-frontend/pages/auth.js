import App from "../components/App";
import PageTitle from "../components/PageTitle";
import connection from "../lib/connection";
import { devices } from "../lib/theme";

const ConnectButton = (props) => {
  return (
    <div>
      <a href={`${connection.API_URL}/auth/${props.service}`}>
        <button>
          <span className="icon">
            <i className={`fab fa-${props.service}`}></i>
          </span>
          <span>{`Connect with ${props.service}`}</span>
        </button>
      </a>
      <style jsx>{`
        button {
          margin: 0px 35px;
          min-width: 250px;
        }
        span {
          text-align: center;
          width: 100%;
        }
      `}</style>
    </div>
  );
};

export default (props) => (
  <App>
    <PageTitle subtitle="Sign Up or Log In" />
    <div className="section">
      <div className="button-row">
        <ConnectButton service="github" />
        <aside>OR</aside>
        <ConnectButton service="google" />
      </div>
    </div>
    <style jsx>{`
      div {
        align-items: center;
        display: flex;
        margin: 0px;
        margin-bottom: 30px;
        text-align: center;
      }
      div.section {
        flex-direction: column;
        flex-grow: 1;
        justify-content: center;
        padding: 0 10px;
      }
      h1, p {
        margin: 0px;
      }
      @media ${devices.tabletOrLarger} {
        h1 {
          font-size: 3rem;
        }
      }
      @media ${devices.smallerThanTablet} {
        h1 {
          font-size: 1.5rem;
        }
      }
      /* Button row */
      .button-row {
        display: flex;
        align-items: stretch;
        justify-content: center;
        align-items: center;
      }
      @media ${devices.smallerThanTablet} {
        .button-row {
          flex-wrap: wrap;
        }
        .button-row * {
          width: 100%;
        }
        .button-row aside {
          margin: 10px 0;
        }
      }
    `}</style>
  </App>
);
