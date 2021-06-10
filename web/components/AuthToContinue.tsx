import { Button, Grid, Typography } from "@material-ui/core";
import { makeStyles } from "@material-ui/core/styles";
import { setRedirectAfterAuth } from "lib/authRedirect";
import { useRouter } from "next/router";
import { FC } from "react";
import { NakedLink } from "./Link";

const useStyles = makeStyles((theme) => ({
  container: {
    marginTop: theme.spacing(12),
  },
}));

export type AuthToContinueProps = {
  label: string;
};

const AuthToContinue: FC<AuthToContinueProps> = ({ label }) => {
  const classes = useStyles();
  const router = useRouter();
  return (
    <Grid container direction="column" alignItems="center" spacing={2} className={classes.container}>
      <Grid item>
        <Typography variant="h3" gutterBottom>
          {label}
        </Typography>
      </Grid>
      <Grid item>
        <Button
          component={NakedLink}
          variant="contained"
          href="/"
          color="primary"
          onClick={() => setRedirectAfterAuth(router.pathname, router.query, router.asPath)}
        >
          Sign up / Log in
        </Button>
      </Grid>
    </Grid>
  );
};

export default AuthToContinue;
