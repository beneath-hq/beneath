import { useMutation } from "@apollo/client";
import { useRouter } from "next/router";
import React, { FC } from "react";

import { TableByOrganizationProjectAndName_tableByOrganizationProjectAndName } from "apollo/types/TableByOrganizationProjectAndName";
import { TableInstance } from "components/table/types";
import { Button, DialogActions, DialogContent, DialogContentText, DialogTitle } from "@material-ui/core";
import { DeleteTableInstance, DeleteTableInstanceVariables } from "apollo/types/DeleteTableInstance";
import { DELETE_TABLE_INSTANCE, QUERY_TABLE } from "apollo/queries/table";
import { toURLName } from "lib/names";
import { makeTableAs, makeTableHref } from "./urls";

export interface DeleteInstanceProps {
  table: TableByOrganizationProjectAndName_tableByOrganizationProjectAndName;
  instance: TableInstance;
  setOpenDialogID: (dialogID: "create" | "promote" | "delete" | null) => void;
}

const DeleteInstance: FC<DeleteInstanceProps> = ({ table, instance, setOpenDialogID }) => {
  const router = useRouter();
  const [deleteTableInstance] = useMutation<DeleteTableInstance, DeleteTableInstanceVariables>(DELETE_TABLE_INSTANCE, {
    onCompleted: (data) => {
      if (data?.deleteTableInstance) {
        router.replace(makeTableHref(table), makeTableAs(table));
      }
      setOpenDialogID(null);
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
        <Button color="primary" autoFocus onClick={() => setOpenDialogID(null)}>
          No, go back
        </Button>
        <Button
          color="primary"
          autoFocus
          onClick={() =>
            deleteTableInstance({
              variables: { instanceID: instance.tableInstanceID },
              refetchQueries: [
                {
                  query: QUERY_TABLE,
                  variables: {
                    organizationName: table.project.organization.name,
                    projectName: table.project.name,
                    tableName: table.name,
                  },
                },
              ],
              update: (cache) => {
                // remove the TableInstance from the cache
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
