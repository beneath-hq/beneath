import Button from "@material-ui/core/Button";
import Container from "@material-ui/core/Container";
import Grid from "@material-ui/core/Grid";
import Typography from "@material-ui/core/Typography";
import { withStyles } from "@material-ui/styles";

import Page from "../components/Page";

const styles = (theme) => ({
});

export default withStyles(styles)(({ classes }) => (
  <Page title="Documentation">
    <div className="section">
      <div className="title">
        <h1>Docs</h1>
      </div>
    </div>
  </Page>
));
