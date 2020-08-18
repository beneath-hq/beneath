import { makeStyles, Theme } from "@material-ui/core/styles";
import { FC } from "react";

type VSpaceProps = {
  units?: number;
};

const useStyles = makeStyles<Theme, VSpaceProps>((theme) => ({
  space: ({ units }) => ({
    display: "block",
    paddingTop: theme.spacing(units || 1),
  }),
}));

const VSpace: FC<VSpaceProps> = ({ units }) => {
  const classes = useStyles({ units });
  return <div className={classes.space} />;
};

export default VSpace;
