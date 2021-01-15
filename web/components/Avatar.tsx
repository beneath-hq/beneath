import React, { FC } from "react";

import Avatar, { AvatarProps } from "@material-ui/core/Avatar";
import { makeStyles, Theme, useTheme } from "@material-ui/core/styles";

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
  toolbar: {
    width: theme.spacing(4.5),
    height: theme.spacing(4.5),
    fontSize: theme.typography.pxToRem(15),
    borderRadius: theme.spacing(4),
    boxShadow: `0 0 0 2px ${theme.palette.secondary.dark}`,
    "&:hover": {
      boxShadow: `0 0 0 2px ${theme.palette.primary.main}`,
    },
  },
}));

export interface BetterAvatarProps extends Omit<AvatarProps, "src"> {
  size: "hero" | "list" | "dense-list" | "toolbar";
  label: string;
  src?: string | null;
}

const BetterAvatar: FC<BetterAvatarProps> = ({ size, label, src, ...other }) => {
  const classes = useStyles();
  const theme = useTheme();

  let className;
  if (size === "hero") {
    className = classes.hero;
  } else if (size === "list") {
    className = classes.list;
  } else if (size === "dense-list") {
    className = classes.denseList;
  } else if (size === "toolbar") {
    className = classes.toolbar;
  }

  // generate a fallback backgroundColor when the user doesn't provide a picture
  let backgroundColor;
  if (!src) {
    let hash = 0;
    for (let i = 0; i < label.length; i++) {
      hash += label.charCodeAt(i);
    }
    const idx = hash % theme.palette.rainbow.length;
    backgroundColor = theme.palette.rainbow[idx];
  }

  return (
    <Avatar
      className={className}
      src={src || undefined}
      alt={label}
      {...other}
      style={{ backgroundColor: backgroundColor }}
    >
      {!src && !!label && label.slice(0, 2)}
    </Avatar>
  );
};

export default BetterAvatar;
