import React from "react";
import Container from "@material-ui/core/Container";
import Typography from "@material-ui/core/Typography";
import { withStyles } from "@material-ui/core/styles";

import Page from "../components/Page";

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

const styles = (theme) => ({
});

class Error extends React.Component {
  static getInitialProps({ res, err }) {
    const statusCode = res ? res.statusCode : (err ? err.statusCode : null);
    return { statusCode };
  }

  render() {
    const { classes, message, statusCode } = this.props;
    return (
      <Page title="Error" contentMarginTop="normal">
        <div>
          <Container maxWidth="lg">
            <Typography component="h2" variant="h4" align="center" gutterBottom>
              {message
                ? message
                : makeErrorMessage(statusCode)}
            </Typography>
          </Container>
        </div>
      </Page>
    );
  }
}

export default withStyles(styles)(Error);
