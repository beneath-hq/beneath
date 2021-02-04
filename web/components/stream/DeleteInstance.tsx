import { useMutation } from "@apollo/client";
import { useRouter } from "next/router";
import React, { FC } from "react";

import { StreamByOrganizationProjectAndName_streamByOrganizationProjectAndName } from "apollo/types/StreamByOrganizationProjectAndName";
import { StreamInstance } from "components/stream/types";
import { Button, DialogActions, DialogContent, DialogContentText, DialogTitle } from "@material-ui/core";
import { DeleteStreamInstance, DeleteStreamInstanceVariables } from "apollo/types/DeleteStreamInstance";
import { DELETE_STREAM_INSTANCE, QUERY_STREAM } from "apollo/queries/stream";
import { toURLName } from "lib/names";
import { makeStreamAs, makeStreamHref } from "./urls";

export interface DeleteInstanceProps {
  stream: StreamByOrganizationProjectAndName_streamByOrganizationProjectAndName;
  instance: StreamInstance;
  setOpenDialogID: (dialogID: "create" | "promote" | "delete" | null) => void;
}

const DeleteInstance: FC<DeleteInstanceProps> = ({ stream, instance, setOpenDialogID }) => {
  const router = useRouter();
  const [deleteStreamInstance] = useMutation<DeleteStreamInstance, DeleteStreamInstanceVariables>(
    DELETE_STREAM_INSTANCE,
    {
      onCompleted: (data) => {
        if (data?.deleteStreamInstance) {
          router.replace(makeStreamHref(stream), makeStreamAs(stream));
        }
        setOpenDialogID(null);
      },
    }
  );

  return (
    <>
      <DialogTitle>Are you sure you want to delete this instance?</DialogTitle>
      <DialogContent>
        <DialogContentText>
          This version and all its data will be permenantly deleted. You won't be able to recover the data.
        </DialogContentText>
      </DialogContent>
      <DialogActions>
        <Button color="primary" autoFocus onClick={() => setOpenDialogID(null)}>
          No, go back
        </Button>
        <Button
          color="primary"
          autoFocus
          onClick={() =>
            deleteStreamInstance({
              variables: { instanceID: instance.streamInstanceID },
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
              update: (cache) => {
                // remove the StreamInstance from the cache
                // hack: see "Solution 2" for why we spread the instance object:
                // https://stackoverflow.com/questions/60697214/how-to-fix-index-signature-is-missing-in-type-error
                cache.evict({ id: cache.identify({ ...instance }) });
                cache.gc();
              },
            })
          }
        >
          Yes, I'm sure
        </Button>
      </DialogActions>
    </>
  );
};

export default DeleteInstance;
