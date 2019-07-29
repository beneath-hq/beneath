import React, { FunctionComponent } from "react";
import Avatar, { AvatarProps } from "@material-ui/core/Avatar";
import { makeStyles, Theme } from "@material-ui/core/styles";

const useStyles = makeStyles((theme: Theme) => ({
  hero: {
    width: theme.spacing(8),
    height: theme.spacing(8),
    fontSize: theme.typography.pxToRem(32),
  },
  list: {
    width: 40,
    height: 40,
    fontSize: theme.typography.pxToRem(20),
  },
  denseList: {
    width: theme.spacing(3),
    height: theme.spacing(3),
    marginRight: theme.spacing(1.5),
    fontSize: theme.typography.pxToRem(12),
  },
}));

interface BetterAvatarProps extends AvatarProps {
  size: string;
  label: string;
  src: string;
}

const BetterAvatar: FunctionComponent<BetterAvatarProps> = ({ size, label, src, ...other }) => {
  const classes = useStyles(undefined);

  let className = undefined;
  if (size === "hero") {
    className = classes.hero;
  } else if (size === "list") {
    className = classes.list;
  } else if (size === "dense-list") {
    className = classes.denseList;
  }
  
  return (
    <Avatar className={className} src={src} alt={label} {...other}>
      {!src && !!label && label.slice(0, 2)}
    </Avatar>
  );
};

export default BetterAvatar;
