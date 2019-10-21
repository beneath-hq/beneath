import { makeStyles, Paper, Theme, Typography } from "@material-ui/core";

const useStyles = makeStyles((theme: Theme) => ({
  paper: {
    padding: theme.spacing(2),
  },
}));

export interface SingleIndicatorProps {
  title: string;
  indicator: string;
  note?: string;
}

export const SingleIndicator = ({ title, indicator, note }: SingleIndicatorProps) => {
  const classes = useStyles();
  return (
    <Paper className={classes.paper}>
      <Typography variant="overline">{title}</Typography>
      <Typography variant="h2" component="p">
        {indicator}
      </Typography>
      {note && <Typography variant="overline">{note}</Typography>}
    </Paper>
  );
};

export default SingleIndicator;
