import Page from "../components/Page";
import { Query } from "react-apollo";
import gql from "graphql-tag";

const PING_QUERY = gql`
  query {
    ping
  }
`;

export default () => (
  <Page title="Ping" >
    <div className="section">
      <Query query={PING_QUERY}>
        {({ loading, error, data }) => {
          if (loading) {
            return <pre className="loading-text">Loading...</pre>;
          } else if (error) {
            return <pre className="error-text">Couldn't ping :(</pre>;
          } else {
            return <pre>Ping: {data.ping}</pre>;
          }
        }}
      </Query>
    </div>
    <style jsx>{`
      .loading-text {
        color: red;
      }
      .error-text {
        color: red;
      }
    `}</style>
  </Page>
);
