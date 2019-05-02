import React from "react";
import Page from "../components/Page";

/**
 * Produces a suitable error message for statusCode
 * @param {*} statusCode 
 */
const makeErrorMessage = (statusCode) => {
  if (!statusCode) {
    return "An unknown error occurred";
  }

  if (statusCode == 404) {
    return "Sorry, but we couldn't find that page";
  }

  if (statusCode == 401) {
    return "Please sign in to view this page";
  }

  return `An error with status ${this.props.statusCode} occurred`;
};

/**
 * Nextjs page that displays an error message
 */
export default class Error extends React.Component {
  static getInitialProps({ res, err }) {
    const statusCode = res ? res.statusCode : (err ? err.statusCode : null);
    return { statusCode };
  }

  render() {
    return (
      <Page title="Error">
        <div>
          <p>
            {this.props.message
              ? this.props.message
              : makeErrorMessage(this.props.statusCode)}
          </p>
        </div>
        <style jsx>{`
          p {
            margin-top: 50px;
            text-align: center;
          }
        `}</style>
      </Page>
    );
  }
}
