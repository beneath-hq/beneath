import { useQuery } from "@apollo/client";
import { Button, ButtonGroup, Dialog, DialogContent, makeStyles, Menu, MenuItem } from "@material-ui/core";
import { MoreVert } from "@material-ui/icons";
import { QUERY_STREAM_INSTANCES } from "apollo/queries/stream";
import { StreamInstanceByOrganizationProjectStreamAndVersion_streamInstanceByOrganizationProjectStreamAndVersion_stream } from "apollo/types/StreamInstanceByOrganizationProjectStreamAndVersion";
import {
  StreamInstancesByOrganizationProjectAndStreamName,
  StreamInstancesByOrganizationProjectAndStreamNameVariables,
} from "apollo/types/StreamInstancesByOrganizationProjectAndStreamName";
import { useRouter } from "next/router";
import { FC, useState } from "react";
import CreateInstance from "./CreateInstance";
import DeleteInstance from "./DeleteInstance";
import PromoteInstance from "./PromoteInstance";
import { StreamInstance } from "./types";
import { makeStreamAs, makeStreamHref } from "./urls";

const useStyles = makeStyles((theme) => ({
  buttonGroup: {
    height: "28px",
  },
  leftPanel: {
    fontSize: "14px",
    "&:hover": {
      cursor: "default",
      backgroundColor: theme.palette.background.default,
    },
    fontWeight: "normal",
    padding: "0px 8px",
  },
  middleButton: {
    padding: "0 9px",
    fontSize: "14px",
    width: "110px",
  },
  rightButton: {
    minWidth: "0",
    padding: "5px 2px",
  },
  icon: {
    fontSize: 20,
  },
}));

interface Props {
  stream: StreamInstanceByOrganizationProjectStreamAndVersion_streamInstanceByOrganizationProjectStreamAndVersion_stream;
  currentInstance: StreamInstance | null;
}

const StreamInstanceSelector: FC<Props> = ({ stream, currentInstance }) => {
  const classes = useStyles();
  const router = useRouter();
  const [openDialogID, setOpenDialogID] = useState<null | "create" | "promote" | "delete">(null);

  // Menu 1: Select a version
  const [menuAnchorEl1, setMenuAnchorEl1] = useState(null);
  const isMenuOpen1 = !!menuAnchorEl1;
  const openMenu1 = (event: any) => setMenuAnchorEl1(event.currentTarget);
  const closeMenu1 = () => setMenuAnchorEl1(null);

  // Menu 2: Create, delete, promote an instance
  const [menuAnchorEl2, setMenuAnchorEl2] = useState(null);
  const isMenuOpen2 = !!menuAnchorEl2;
  const openMenu2 = (event: any) => setMenuAnchorEl2(event.currentTarget);
  const closeMenu2 = () => setMenuAnchorEl2(null);

  const organizationName = stream.project.organization.name;
  const projectName = stream.project.name;
  const streamName = stream.name;

  const { error, data } = useQuery<
    StreamInstancesByOrganizationProjectAndStreamName,
    StreamInstancesByOrganizationProjectAndStreamNameVariables
  >(QUERY_STREAM_INSTANCES, {
    variables: {
      organizationName,
      projectName,
      streamName,
    },
  });

  if (error || !data) {
    return null;
  }

  const instances = data.streamInstancesByOrganizationProjectAndStreamName;

  // create actions for dropdown menu #2
  const instanceActions = [
    {
      label: "Create instance",
      onClick: () => {
        closeMenu2();
        setOpenDialogID("create");
      },
    },
  ];
  if (currentInstance && currentInstance.streamInstanceID !== stream.primaryStreamInstanceID) {
    instanceActions.push({
      label: "Promote to primary",
      onClick: () => {
        closeMenu2();
        setOpenDialogID("promote");
      },
    });
  }
  if (currentInstance) {
    instanceActions.push({
      label: "Delete instance",
      onClick: () => {
        closeMenu2();
        setOpenDialogID("delete");
      },
    });
  }

  return (
    <>
      <ButtonGroup disableElevation className={classes.buttonGroup}>
        <Button className={classes.leftPanel}>Version</Button>
        <Button onClick={openMenu1} className={classes.middleButton}>
          {`${currentInstance?.version !== undefined ? currentInstance.version : ""}` +
            (currentInstance?.streamInstanceID === stream.primaryStreamInstanceID ? " (Primary)" : "")}
        </Button>
        {stream.project.permissions.create && stream.allowManualWrites && !stream.meta && (
          <Button className={classes.rightButton} onClick={openMenu2}>
            <MoreVert className={classes.icon} />
          </Button>
        )}
      </ButtonGroup>

      {/* MENU 1 (list of instances) */}
      <Menu
        autoFocus={false}
        anchorEl={menuAnchorEl1}
        open={isMenuOpen1}
        onClose={closeMenu1}
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
        {instances.map((instance, idx) => (
          <MenuItem
            key={idx}
            onClick={() => {
              closeMenu1();
              router.replace(makeStreamHref(stream, instance), makeStreamAs(stream, instance));
            }}
          >
            {`${instance.version}` + (instance.streamInstanceID === stream.primaryStreamInstanceID ? " (Primary)" : "")}
          </MenuItem>
        ))}
      </Menu>

      {/* MENU 2 (instance action items) */}
      <Menu
        autoFocus={false}
        anchorEl={menuAnchorEl2}
        open={isMenuOpen2}
        onClose={closeMenu2}
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
        {instanceActions.map((action, idx) => (
          <MenuItem key={idx} onClick={action.onClick}>
            {action.label}
          </MenuItem>
        ))}
      </Menu>

      <Dialog open={openDialogID === "create"} onBackdropClick={() => setOpenDialogID(null)}>
        <DialogContent>
          <CreateInstance stream={stream} instances={instances} setOpenDialogID={setOpenDialogID} />
        </DialogContent>
      </Dialog>
      {currentInstance && (
        <>
          <Dialog open={openDialogID === "promote"} onBackdropClick={() => setOpenDialogID(null)}>
            <DialogContent>
              <PromoteInstance stream={stream} instance={currentInstance} setOpenDialogID={setOpenDialogID} />
            </DialogContent>
          </Dialog>
          <Dialog open={openDialogID === "delete"} onBackdropClick={() => setOpenDialogID(null)}>
            <DialogContent>
              <DeleteInstance stream={stream} instance={currentInstance} setOpenDialogID={setOpenDialogID} />
            </DialogContent>
          </Dialog>
        </>
      )}
    </>
  );
};

export default StreamInstanceSelector;
