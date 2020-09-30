import React, { FC } from "react";
import Button from "@material-ui/core/Button";
import ButtonGroup from "@material-ui/core/ButtonGroup";
import ArrowDropDownIcon from "@material-ui/icons/ArrowDropDown";
import Menu from "@material-ui/core/Menu";
import MenuItem from "@material-ui/core/MenuItem";

import { NakedLink } from "./Link";
import { makeStyles, PropTypes } from "@material-ui/core";
import clsx from "clsx";

const useStyles = makeStyles((_) => ({
  denseButton: {
    minWidth: "0",
    padding: "5px 8px",
  },
  denseRightButton: {
    minWidth: "0",
    padding: "5px 2px",
  },
}));

export interface SplitButtonProps {
  className?: string;
  color?: PropTypes.Color;
  margin?: "normal" | "dense";
  variant?: "text" | "outlined" | "contained";
  actions: { label: string; href: string; as?: string }[];
  mainActionIdx?: number;
  removeMainFromDropdown?: boolean;
}

export const SplitButton: FC<SplitButtonProps> = ({
  className,
  color,
  margin,
  variant,
  actions,
  mainActionIdx,
  removeMainFromDropdown,
}) => {
  const [menuAnchorEl, setMenuAnchorEl] = React.useState(null);
  const isMenuOpen = !!menuAnchorEl;
  const openMenu = (event: any) => setMenuAnchorEl(event.currentTarget);
  const closeMenu = () => setMenuAnchorEl(null);

  mainActionIdx = mainActionIdx || 0;
  const mainAction = actions[mainActionIdx];

  const classes = useStyles();
  return (
    <>
      <ButtonGroup className={className} disableElevation variant={variant} color={color} aria-label="split button">
        <Button
          className={clsx(margin === "dense" && classes.denseButton)}
          component={NakedLink}
          href={mainAction.href}
          as={mainAction.href}
        >
          {mainAction.label}
        </Button>
        <Button
          className={clsx(margin === "dense" && classes.denseRightButton)}
          size="small"
          aria-expanded={isMenuOpen ? "true" : undefined}
          aria-haspopup="menu"
          onClick={openMenu}
        >
          <ArrowDropDownIcon />
        </Button>
      </ButtonGroup>
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
        {actions.map(
          (action, idx) =>
            (!removeMainFromDropdown || idx !== mainActionIdx) && (
              <MenuItem key={idx} component={NakedLink} href={action.href} as={action.as}>
                {action.label}
              </MenuItem>
            )
        )}
      </Menu>
    </>
  );
};

export default SplitButton;
