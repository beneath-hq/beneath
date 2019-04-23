import Button from "@material-ui/core/Button";
import Grid from "@material-ui/core/Grid";
import Typography from "@material-ui/core/Typography";
import { withStyles } from '@material-ui/core/styles';

import Page from "../components/Page";

const styles = (theme) => ({
  heroContainer: {
    maxWidth: 900,
    margin: '0 auto',
    padding: `${theme.spacing.unit * 8}px 0 ${theme.spacing.unit * 6}px`,
  },
  heroButtonsContainer: {
    marginTop: theme.spacing.unit * 2,
  },
});

export default withStyles(styles)(({ classes }) => (
  <Page title="Home">
    <div className={classes.heroContainer}>
      <Typography component="h1" variant="h4" align="center" gutterBottom>
        Data Science for the Decentralised Economy
      </Typography>
      <Typography component="h2" variant="subtitle1" align="center" gutterBottom>
        Beneath is a full Ethereum data science platform. Explore other people's analytics or start building your own.
      </Typography>
      <div className={classes.heroButtonsContainer}>
        <Grid container spacing={16} justify="center">
          <Grid item>
            <Button size="large" color="primary" variant="outlined" href="https://network.us18.list-manage.com/subscribe?u=ead8c956abac88f03b662cf03&id=466a1d05d9">
              Get the newsletter
            </Button>
          </Grid>
          <Grid item>
            <Button size="large" color="primary" variant="outlined" href="mailto:contact@beneath.network">
              Get in touch
            </Button>    
          </Grid>
        </Grid>
      </div>
    </div>
  </Page>
));
