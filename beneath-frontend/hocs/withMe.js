import { Query } from "react-apollo";

import { QUERY_ME } from "../queries/user";

const withMe = (Component) => {
  return (props) => (
    <Query query={QUERY_ME}>
      {({ loading, error, data }) => {
        if (data) {
          let { me } = data;
          return <Component {...props} me={me} />
        } else if (error) {
          console.log("withMe error: ", error);
        }
        return null;
      }}
    </Query>
  );
}

export default withMe;
