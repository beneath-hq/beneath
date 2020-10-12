import clsx from "clsx";
import React, { FC } from "react";
import Button from "@material-ui/core/Button";
import Menu from "@material-ui/core/Menu";
import MenuItem from "@material-ui/core/MenuItem";
import { makeStyles, PropTypes } from "@material-ui/core";

import { NakedLink } from "./Link";

const useStyles = makeStyles((_) => ({
  denseButton: {
    minWidth: "0",
    padding: "6px 6px",
  },
}));

export interface DropdownButtonProps {
  className?: string;
  color?: PropTypes.Color;
  margin?: "normal" | "dense";
  variant?: "text" | "outlined" | "contained";
  actions: { label: string; href?: string; as?: string; onClick?: () => void; }[];
  children?: any;
}

export const DropdownButton: FC<DropdownButtonProps> = ({ className, color, margin, variant, actions, children }) => {
  const [menuAnchorEl, setMenuAnchorEl] = React.useState(null);
  const isMenuOpen = !!menuAnchorEl;
  const openMenu = (event: any) => setMenuAnchorEl(event.currentTarget);
  const closeMenu = () => setMenuAnchorEl(null);

  const classes = useStyles();
  return (
    <>
      <Button
        className={clsx(className, margin === "dense" && classes.denseButton)}
        variant={variant}
        color={color}
        aria-expanded={isMenuOpen ? "true" : undefined}
        aria-haspopup="menu"
        onClick={openMenu}
      >
        {children}
      </Button>
      <Menu
        autoFocus={false}
        anchorEl={menuAnchorEl}
        open={isMenuOpen}
        onClose={closeMenu}
        anchorOrigin={{
          vertical: "bottom",
          horizontal: "right",
        }}
        transformOrigin={{
          vertical: "top",
          horizontal: "right",
        }}
        getContentAnchorEl={null}
      >
        {actions.map((action, idx) => (
          <MenuItem key={idx}
            onClick={() => {
              if (action.onClick) {
                action.onClick();
                closeMenu();
              }
            }}
            component={action.href ? NakedLink : "li"} href={action.href} as={action.as}>
            {action.label}
          </MenuItem>
        ))}
      </Menu>
    </>
  );
};

export default DropdownButton;
