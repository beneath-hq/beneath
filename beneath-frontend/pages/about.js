import App from "../components/App";
import PageTitle from "../components/PageTitle";
import { devices } from "../lib/theme";

const Button = props => {
  return (
    <div>
      <a href={props.href}>
        <button>
          <span>{props.text}</span>
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

export default () => (
  <App>
    <PageTitle subtitle="Home" />
    <div className="section">
      <div className="title">
        <h1>Data Science for the Decentralised Economy</h1>
      </div>
      <div className="bullets">
        <p>
          Beneath is a full Ethereum data science platform. Explore other people's analytics or start building your own.
        </p>
      </div>
      <div className="button-row">
        <Button
          href="https://network.us18.list-manage.com/subscribe?u=ead8c956abac88f03b662cf03&id=466a1d05d9"
          text="Get the newsletter"
        />
        <aside>OR</aside>
        <Button href="mailto:contact@beneath.network" text="Get in touch" />
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
