import { useMutation } from "@apollo/client";
import { Button, DialogActions, DialogContent, DialogContentText, DialogTitle } from "@material-ui/core";
import React, { FC } from "react";

import { QUERY_STREAM, UPDATE_STREAM_INSTANCE } from "apollo/queries/stream";
import { StreamByOrganizationProjectAndName_streamByOrganizationProjectAndName } from "apollo/types/StreamByOrganizationProjectAndName";
import { UpdateStreamInstance, UpdateStreamInstanceVariables } from "apollo/types/UpdateStreamInstance";
import { StreamInstance } from "components/stream/types";

export interface PromoteInstanceProps {
  stream: StreamByOrganizationProjectAndName_streamByOrganizationProjectAndName;
  instance: StreamInstance;
  setInstance: (instance: StreamInstance | null) => void;
  setOpenDialogID: (dialogID: "create" | "promote" | "delete" | null) => void;
}

const PromoteInstance: FC<PromoteInstanceProps> = ({ stream, instance, setInstance, setOpenDialogID }) => {
  const [updateStreamInstance] = useMutation<UpdateStreamInstance, UpdateStreamInstanceVariables>(
    UPDATE_STREAM_INSTANCE,
    {
      onCompleted: (data) => {
        if (data?.updateStreamInstance) {
          setInstance(data.updateStreamInstance);
        }
        setOpenDialogID(null);
      },
    }
  );

  return (
    <>
      <DialogTitle>Are you sure you want to make this instance the primary instance?</DialogTitle>
      <DialogContent>
        <DialogContentText>
          All prior versions and their data will be deleted. You won't be able to recover the data.
        </DialogContentText>
      </DialogContent>
      <DialogActions>
        <Button color="primary" autoFocus onClick={() => setOpenDialogID(null)}>
          No, go back
        </Button>
        <Button
          color="primary"
          autoFocus
          onClick={() => {
            updateStreamInstance({
              variables: { input: { streamInstanceID: instance.streamInstanceID, makePrimary: true } },
              refetchQueries: [
                {
                  query: QUERY_STREAM,
                  variables: {
                    organizationName: stream.project.organization.name,
                    projectName: stream.project.name,
                    streamName: stream.name,
                  },
                },
              ],
            });
            setOpenDialogID(null);
          }}
        >
          Yes, I'm sure
        </Button>
      </DialogActions>
    </>
  );
};

export default PromoteInstance;
