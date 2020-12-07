import { useQuery } from "@apollo/client";
import { Chip, Grid, Typography, makeStyles, Dialog, DialogContent } from "@material-ui/core";
import { MoreVert } from "@material-ui/icons";
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
import { toURLName } from "lib/names";

const useStyles = makeStyles((theme) => ({
  container: {
    marginTop: theme.spacing(2),
    marginBottom: theme.spacing(2),
  },
  streamName: {
    fontSize: theme.typography.pxToRem(36),
    fontWeight: "bold",
  },
  site: {
    display: "block",
  },
  selectField: {
    width: "170px",
  },
  dropdownButton: {
    backgroundColor: theme.palette.background.paper,
    "&:hover": {
      backgroundColor: theme.palette.secondary.main,
    },
    border: `1px solid ${theme.palette.border.paper}`,
    color: theme.palette.common.white,
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
  const classes = useStyles();
  const metrics = useMonthlyMetrics(EntityKind.Stream, stream.streamID).total;
  const [instances, setInstances] = useState<Instance[]>([]);
  const organizationName = stream.project.organization.name;
  const projectName = stream.project.name;
  const streamName = stream.name;

  const { error, data } = useQuery<
    StreamInstancesByOrganizationProjectAndStreamName,
    StreamInstancesByOrganizationProjectAndStreamNameVariables
  >(QUERY_STREAM_INSTANCES, {
    variables: { organizationName, projectName, streamName },
  });
  if (error) {
    console.error("Unexpected error loading instances: ", error);
  }

  useEffect(() => {
    if (data?.streamInstancesByOrganizationProjectAndStreamName && data.streamInstancesByOrganizationProjectAndStreamName.length > 0 ) {
      // sort instances by version (so the SelectField shows them in a sensible order)
      const instances: Instance[] = [];
      instances.push(...data.streamInstancesByOrganizationProjectAndStreamName);
      instances.sort((a, b) => a.version < b.version ? 1 : -1);
      setInstances(instances);

      // if there's no set instance (which happens when there's no primary instance), show the instance with the highest version
      if (!instance && instances.length > 0) {
        setInstance(instances[0]);
      }
    }
  }, [data?.streamInstancesByOrganizationProjectAndStreamName]);

  const instanceActions = [{ label: "Create instance", onClick: () => setOpenDialogID("create")}];
  if (instance && !instance?.madePrimaryOn) {
    instanceActions.push({ label: "Promote to primary", onClick: () => setOpenDialogID("promote") });
  }
  if (instance) {
    instanceActions.push({ label: "Delete instance", onClick: () => setOpenDialogID("delete") });
  }

  return (
    <Grid container justify="space-between" alignItems="flex-start" spacing={4} className={classes.container}>
      <Grid item>
        <Grid container direction="column" spacing={1}>
          <Grid item>
            <Grid container alignItems="center" spacing={2}>
              <Grid item>
                <Typography className={classes.streamName}>{toURLName(streamName)}</Typography>
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
                <Grid container spacing={1} alignItems="center">
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
                      onChange={(_, value) => {
                        if (value) {
                          setInstance(value as Instance);
                        }
                      }}
                      margin="none"
                    />
                  </Grid>
                  <Grid item>
                    <DropdownButton
                      variant="contained"
                      margin="dense"
                      actions={instanceActions}
                      className={classes.dropdownButton}
                    >
                      <MoreVert />
                    </DropdownButton>
                  </Grid>
                  <Dialog open={openDialogID === "create"} onBackdropClick={() => setOpenDialogID(null)}>
                    <DialogContent>
                      <CreateInstance
                        stream={stream}
                        instances={instances}
                        setInstance={setInstance}
                        setOpenDialogID={setOpenDialogID}
                      />
                    </DialogContent>
                  </Dialog>
                  {instance && (
                    <>
                      <Dialog open={openDialogID === "promote"} onBackdropClick={() => setOpenDialogID(null)}>
                        <DialogContent>
                          <PromoteInstance
                            stream={stream}
                            instance={instance}
                            setInstance={setInstance}
                            setOpenDialogID={setOpenDialogID}
                          />
                        </DialogContent>
                      </Dialog>
                      <Dialog open={openDialogID === "delete"} onBackdropClick={() => setOpenDialogID(null)}>
                        <DialogContent>
                          <DeleteInstance
                            stream={stream}
                            instance={instance}
                            instances={instances}
                            setInstance={setInstance}
                            setOpenDialogID={setOpenDialogID}
                          />
                        </DialogContent>
                      </Dialog>
                    </>
                  )}
                </Grid>
              </Grid>
            </Grid>
          </Grid>
          <Grid item>
            <Typography variant="body1">{stream.description}</Typography>
          </Grid>
        </Grid>
      </Grid>
      <Grid item>
        <Chip
          label={`${prettyPrintBytes(metrics.writeBytes)} written \xa0\xa0\ ${prettyPrintBytes(
            metrics.readBytes
          )} read`}
          clickable
          component={NakedLink}
          href={`/stream?organization_name=${organizationName}&project_name=${projectName}&stream_name=${streamName}&tab=monitoring`}
          as={`/${organizationName}/${projectName}/${streamName}/-/monitoring`}
        />
      </Grid>
    </Grid>
  );
};

export default StreamHero;
