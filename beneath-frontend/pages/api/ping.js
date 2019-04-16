import Page from "../../components/Page";
import { Query } from "react-apollo";
import gql from "graphql-tag";

// Queries
const PING_QUERY = gql`
  {
    ping
  }
`;

export default props => (
  <Page>
    <article>
      <Query query={PING_QUERY}>
        {({ loading, error, data }) => {
          if (loading) {
            return <pre className="loading-text">Loading...</pre>;
          } else if (error) {
            return <pre className="error-text">Couldn't load API key</pre>;
          } else {
            return <p>Ping: {data.ping}</p>;
          }
        }}
      </Query>
      <p>Remember to keep it safe!</p>
      <style jsx>{`
        .loading-text {
          color: red;
        }
        .error-text {
          color: red;
        }
      `}</style>
    </article>
  </Page>
);
