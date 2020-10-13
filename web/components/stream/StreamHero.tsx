import { useQuery } from "@apollo/client";
import { Chip, Grid, Typography, makeStyles, Dialog, DialogContent, DialogTitle, DialogContentText, DialogActions, Button } from "@material-ui/core";
import { ArrowDropDown, Settings } from "@material-ui/icons";
import { FC, useEffect, useState } from "react";

import { NakedLink } from "../Link";
import { useMonthlyMetrics } from "../metrics/hooks";
import { prettyPrintBytes } from "../metrics/util";
import SelectField from "components/forms/SelectField";
import { EntityKind } from "apollo/types/globalTypes";
import { QUERY_STREAM_INSTANCES } from "apollo/queries/stream";
import {
  StreamByOrganizationProjectAndName_streamByOrganizationProjectAndName,
} from "apollo/types/StreamByOrganizationProjectAndName";
import {
  StreamInstancesByOrganizationProjectAndStreamNameVariables,
  StreamInstancesByOrganizationProjectAndStreamName,
} from "apollo/types/StreamInstancesByOrganizationProjectAndStreamName";
import DropdownButton from "components/DropdownButton";
import CreateInstance from "./CreateInstance";
import DeleteInstance from "./DeleteInstance";
import PromoteInstance from "./PromoteInstance";
import { Instance } from "pages/stream";

const useStyles = makeStyles((theme) => ({
  container: {
    marginTop: theme.spacing(2),
    marginBottom: theme.spacing(2),
  },
  site: {
    display: "block",
  },
  selectField: {
    width: "170px",
  },
  dropdownButton: {
    marginTop: theme.spacing(.9)
  }
}));

export interface StreamHeroProps {
  stream: StreamByOrganizationProjectAndName_streamByOrganizationProjectAndName;
  instance: Instance | null;
  setInstance: (instance: Instance | null) => void;
  openDialogID: string | null;
  setOpenDialogID: (dialogID: "create" | "promote" | "delete" | null) => void;
}

const StreamHero: FC<StreamHeroProps> = ({ stream, instance, setInstance, openDialogID, setOpenDialogID }) => {
  const [firstLoad, setFirstLoad] = useState(true);
  const organizationName = stream.project.organization.name;
  const projectName = stream.project.name;
  const streamName = stream.name;

  const metrics = useMonthlyMetrics(EntityKind.Stream, stream.streamID).total;

  const { error, data } = useQuery<
    StreamInstancesByOrganizationProjectAndStreamName,
    StreamInstancesByOrganizationProjectAndStreamNameVariables
  >(QUERY_STREAM_INSTANCES, {
    variables: { organizationName, projectName, streamName },
  });
  if (error) {
    console.error("Unexpected error loading instances: ", error);
  }

  const instances: Instance[] = [];
  if (data) {
    instances.push(...data.streamInstancesByOrganizationProjectAndStreamName);
    instances.sort((a, b) => a.version < b.version ? 1 : -1);
  }

  useEffect(() => {
    if (instances && !stream.primaryStreamInstanceID && firstLoad) {
      setInstance(instances[0] ? instances[0] : null);
      setFirstLoad(false);
    }
  }, [instances]);

  const instanceActions = [{ label: "Create instance", onClick: () => setOpenDialogID("create")}];
  if (instance && !instance?.madePrimaryOn) {
    instanceActions.push({ label: "Promote to primary", onClick: () => setOpenDialogID("promote") });
  }
  if (instance) {
    instanceActions.push({ label: "Delete instance", onClick: () => setOpenDialogID("delete") });
  }

  const classes = useStyles();
  return (
    <Grid container justify="space-between" alignItems="flex-start" spacing={4} className={classes.container}>
      <Grid item>
        <Grid container direction="column" spacing={1}>
          <Grid item>
            <Grid container alignItems="center" spacing={2}>
              <Grid item>
                <Typography component="h1" variant="h1">
                  {streamName}
                </Typography>
              </Grid>
              <Grid item>
                <Chip
                  label={stream.project.public ? "Public" : "Private"}
                  clickable
                  component={NakedLink}
                  href={`/project?organization_name=${organizationName}&project_name=${projectName}&tab=members`}
                  as={`/${organizationName}/${projectName}/-/members`}
                />
              </Grid>
              <Grid item>
                <Chip
                  label={prettyPrintBytes(metrics.writeBytes) + " written"}
                  clickable
                  component={NakedLink}
                  href={`/stream?organization_name=${organizationName}&project_name=${projectName}&stream_name=${streamName}&tab=monitoring`}
                  as={`/${organizationName}/${projectName}/${streamName}/-/monitoring`}
                />
              </Grid>
              <Grid item>
                <Chip
                  label={prettyPrintBytes(metrics.readBytes) + " read"}
                  clickable
                  component={NakedLink}
                  href={`/stream?organization_name=${organizationName}&project_name=${projectName}&stream_name=${streamName}&tab=monitoring`}
                  as={`/${organizationName}/${projectName}/${streamName}/-/monitoring`}
                />
              </Grid>
            </Grid>
          </Grid>
          <Grid item>
            <Typography variant="body1">{stream.description}</Typography>
          </Grid>
        </Grid>
      </Grid>
      <Grid item>
        <Grid container spacing={1}>
          <Grid item className={classes.selectField}>
            <SelectField
              id="instanceID"
              required
              options={instances}
              getOptionLabel={(option: Instance) => {
                const versionString = `v${option.version.toString()}`;
                const primaryTag = option.madePrimaryOn ? " (primary)" : "";
                const finalTag = option.madeFinalOn ? " (final)" : "";
                return versionString + primaryTag + finalTag;
              }}
              getOptionSelected={(option: Instance, value: Instance) => {
                return option.version === value.version;
              }}
              value={instance}
              multiple={false}
              onChange={( _, value ) => {
                if (value) {
                  setInstance(value as Instance);
                }
              }}
              margin="none"
            />
          </Grid>
          <Grid item>
            <DropdownButton
              // className={clsx(classes.rightItem, classes.rightButton)}
              // color="secondary"
              variant="contained"
              margin="dense"
              actions={instanceActions}
              className={classes.dropdownButton}
            >
              <Settings />
              <ArrowDropDown />
            </DropdownButton>
          </Grid>
          <Dialog open={openDialogID === "create"} onBackdropClick={() => setOpenDialogID(null)}>
            <DialogContent>
              <CreateInstance stream={stream} instances={instances} setInstance={setInstance} setOpenDialogID={setOpenDialogID}/>
            </DialogContent>
          </Dialog>
          {instance && (
            <>
              <Dialog open={openDialogID === "promote"} onBackdropClick={() => setOpenDialogID(null)}>
                <DialogContent>
                  <PromoteInstance stream={stream} instance={instance} setInstance={setInstance} setOpenDialogID={setOpenDialogID}/>
                </DialogContent>
              </Dialog>
              <Dialog open={openDialogID === "delete"} onBackdropClick={() => setOpenDialogID(null)}>
                <DialogContent>
                  <DeleteInstance stream={stream} instance={instance} instances={instances} setInstance={setInstance} setOpenDialogID={setOpenDialogID}/>
                </DialogContent>
              </Dialog>
            </>
          )}
        </Grid>
      </Grid>
    </Grid>
  );
};

export default StreamHero;
