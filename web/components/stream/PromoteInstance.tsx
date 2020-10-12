import { useMutation } from "@apollo/client";
import React, { FC } from "react";

import { StreamByOrganizationProjectAndName_streamByOrganizationProjectAndName } from "apollo/types/StreamByOrganizationProjectAndName";
import { Instance } from "pages/stream";
import { Button, DialogActions, DialogContent, DialogContentText, DialogTitle } from "@material-ui/core";
import { StageStreamInstance, StageStreamInstanceVariables } from "apollo/types/StageStreamInstance";
import { QUERY_STREAM_INSTANCES, STAGE_STREAM_INSTANCE } from "apollo/queries/stream";

export interface PromoteInstanceProps {
  stream: StreamByOrganizationProjectAndName_streamByOrganizationProjectAndName;
  instance: Instance;
  setInstance: (instance: Instance | null) => void;
  setOpenDialogID: (dialogID: "create" | "promote" | "delete" | null) => void;
}

const PromoteInstance: FC<PromoteInstanceProps> = ({ stream, instance, setInstance, setOpenDialogID }) => {
  const [stageStreamInstance] = useMutation<StageStreamInstance, StageStreamInstanceVariables>(STAGE_STREAM_INSTANCE, {
    onCompleted: (data) => {
      if (data?.stageStreamInstance) {
        setInstance(data.stageStreamInstance);
      }
      setOpenDialogID(null);
    },
  });

  return (
    <>
      <DialogTitle>Are you sure you want to make this instance the primary instance?</DialogTitle>
      <DialogContent>
        <DialogContentText>
          All prior versions and their data will be deleted. You won't be able to recover the data.
        </DialogContentText>
      </DialogContent>
      <DialogActions>
        <Button
          color="primary"
          autoFocus
          onClick={() => setOpenDialogID(null)}
        >
          No, go back
          </Button>
        <Button
          color="primary"
          autoFocus
          onClick={() => {
            stageStreamInstance({
              variables: {streamID: stream.streamID, version: instance.version, makePrimary: true},
              update: (cache, { data }) => {
                if (data && data.stageStreamInstance) {
                  const queryData = cache.readQuery({
                    query: QUERY_STREAM_INSTANCES,
                    variables: { organizationName: stream.project.organization.name, projectName: stream.project.name, streamName: stream.name },
                  }) as any;

                  const filtered = queryData.streamInstancesByOrganizationProjectAndStreamName.filter(
                    (instnc: Instance) => instnc.version >= instance.version
                  );

                  cache.writeQuery({
                    query: QUERY_STREAM_INSTANCES,
                    variables: { organizationName: stream.project.organization.name, projectName: stream.project.name, streamName: stream.name },
                    data: { streamInstancesByOrganizationProjectAndStreamName: filtered },
                  });
                }
              },
            });
            setOpenDialogID(null);
            }
          }
        >
          Yes, I'm sure
          </Button>
      </DialogActions>
    </>
  );
};

export default PromoteInstance;
