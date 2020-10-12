import { useMutation } from "@apollo/client";
import { useRouter } from "next/router";
import React, { FC } from "react";

import useMe from "../../hooks/useMe";
import { StreamByOrganizationProjectAndName_streamByOrganizationProjectAndName } from "apollo/types/StreamByOrganizationProjectAndName";
import { Instance } from "pages/stream";
import { Button, DialogActions, DialogContent, DialogContentText, DialogTitle } from "@material-ui/core";
import { DeleteStreamInstance, DeleteStreamInstanceVariables } from "apollo/types/DeleteStreamInstance";
import { DELETE_STREAM_INSTANCE, QUERY_STREAM_INSTANCES } from "apollo/queries/stream";

export interface DeleteInstanceProps {
  stream: StreamByOrganizationProjectAndName_streamByOrganizationProjectAndName;
  instance: Instance;
  instances: Instance[];
  setInstance: (instance: Instance | null) => void;
  setOpenDialogID: (dialogID: "create" | "promote" | "delete" | null) => void;
}

const DeleteInstance: FC<DeleteInstanceProps> = ({ stream, instance, instances, setInstance, setOpenDialogID }) => {
  const me = useMe();
  const router = useRouter();
  const [deleteStreamInstance] = useMutation<DeleteStreamInstance, DeleteStreamInstanceVariables>(DELETE_STREAM_INSTANCE, {
    onCompleted: (data) => {
      if (data?.deleteStreamInstance) {
        // go to the instance w/ highest version #
        setInstance(instances[0] ? instances[0] : null);
      }
    },
  });

  return (
    <>
      <DialogTitle>Are you sure you want to delete this instance?</DialogTitle>
        <DialogContent>
          <DialogContentText>
          This version and all its data will be permenantly deleted. You won't be able to recover the data.
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
              deleteStreamInstance({
                variables: { instanceID: instance.streamInstanceID },
                update: (cache, {data}) => {
                  if (data && data.deleteStreamInstance) {
                    const queryData = cache.readQuery({
                      query: QUERY_STREAM_INSTANCES,
                      variables: {organizationName: stream.project.organization.name, projectName: stream.project.name, streamName: stream.name },
                    }) as any;

                    const filtered = queryData.streamInstancesByOrganizationProjectAndStreamName.filter(
                      (instnc: Instance) => instnc.streamInstanceID !== instance.streamInstanceID
                    );

                    cache.writeQuery({
                      query: QUERY_STREAM_INSTANCES,
                      variables: {organizationName: stream.project.organization.name, projectName: stream.project.name, streamName: stream.name },
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

export default DeleteInstance;
