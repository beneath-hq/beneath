import Page from "../../components/Page";
import { Query } from "react-apollo";
import gql from "graphql-tag";

// Queries
const ME_QUERY = gql`
  query {
    me {
      userId
      name
      email
    }
  }
`;

const MyData = props => (
  <div>
    <Query query={ME_QUERY}>
      {({ loading, error, data }) => {
        if (loading) {
          return <pre className="loading-text">Loading...</pre>;
        } else if (error) {
          return (
            <pre className="error-text">
              Couldn't load query: {JSON.stringify(error, null, 2)}
            </pre>
          );
        } else {
          return <pre>{JSON.stringify(data, null, 2)}</pre>;
        }
      }}
    </Query>
    <style jsx>{`
      .loading-text {
        color: red;
      }
      .error-text {
        color: red;
      }
    `}</style>
  </div>
);

export default props => (
  <Page>
    <article>
      <h1>My data</h1>
      <MyData />
      <p>That's what we know about you!</p>
    </article>
  </Page>
);
