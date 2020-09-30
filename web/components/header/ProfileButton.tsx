import React, { FC } from "react";

import { Me_me } from "../../apollo/types/Me";
import UsageIndicator from "../metrics/user/UsageIndicator";
import { NakedLink } from "../Link";

import {
  Button,
  makeStyles,
  Menu,
  MenuItem,
  Typography,
} from "@material-ui/core";

import Avatar from "components/Avatar";
import clsx from "clsx";

const useStyles = makeStyles((_) => ({
  avatarButton: {
    minWidth: "0",
    "&:hover": {
      backgroundColor: "inherit",
    },
  },
  menuPaper: {
    minWidth: "250px",
  },
  menuItemHeader: {
    borderBottom: `1px solid rgba(255, 255, 255, 0.175)`,
  },
  menuItemUsage: {
    borderBottom: `1px solid rgba(255, 255, 255, 0.175)`,
  },
}));

export interface ProfileButtonProps {
  me: Me_me;
  className?: string;
}

export const ProfileButton: FC<ProfileButtonProps> = ({ me, className }) => {
  const [menuAnchorEl, setMenuAnchorEl] = React.useState(null);
  const isMenuOpen = !!menuAnchorEl;
  const openMenu = (event: any) => setMenuAnchorEl(event.currentTarget);
  const closeMenu = () => setMenuAnchorEl(null);

  const makeMenuItem = (text: string | JSX.Element, props: any) => {
    return (
      <MenuItem component={props.as ? NakedLink : "a"} {...props}>
        {text}
      </MenuItem>
    );
  };

  const classes = useStyles();
  return (
    <>
      <Button className={clsx(classes.avatarButton, className)} aria-haspopup="true" onClick={openMenu}>
        <Avatar size="toolbar" label={me.name} src={me.photoURL} />
      </Button>
      <Menu
        autoFocus={false}
        anchorEl={menuAnchorEl}
        open={isMenuOpen}
        onClose={closeMenu}
        PaperProps={{ className: classes.menuPaper }}
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
        <MenuItem disabled className={classes.menuItemHeader}>
          <div>
            <Typography variant="h4">{me.displayName}</Typography>
            <Typography variant="subtitle1" gutterBottom>
              @{me.name}
            </Typography>
          </div>
        </MenuItem>
        {me.readQuota &&
          me.personalUser?.billingOrganizationID === me.organizationID &&
          makeMenuItem(<UsageIndicator standalone={false} kind="read" usage={me.readUsage} quota={me.readQuota} />, {
            onClick: closeMenu,
            as: `/${me.name}/-/monitoring`,
            href: `/organization?organization_name=${me.name}&tab=monitoring`,
            className: classes.menuItemUsage,
          })}
        {makeMenuItem("Profile", {
          onClick: closeMenu,
          as: `/${me.name}`,
          href: `/organization?organization_name=${me.name}`,
        })}
        {makeMenuItem("Secrets", {
          onClick: closeMenu,
          as: `/${me.name}/-/secrets`,
          href: `/organization?organization_name=${me.name}&tab=secrets`,
        })}
        {me.personalUser &&
          me.personalUser.billingOrganizationID !== me.organizationID &&
          makeMenuItem(me.personalUser.billingOrganization.displayName || me.personalUser.billingOrganization.name, {
            onClick: closeMenu,
            as: `/${me.personalUser.billingOrganization.name}`,
            href: `/organization?organization_name=${me.personalUser.billingOrganization.name}`,
          })}
        {makeMenuItem("Logout", {
          href: `/-/redirects/auth/logout`,
        })}
      </Menu>
    </>
  );
};

export default ProfileButton;
